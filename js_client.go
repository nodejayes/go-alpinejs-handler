package goalpinejshandler

import (
	"bytes"
	"fmt"
)

const src = `
window.alpinestorehandler = {};
window.alpinestorehandler.eventEmitter = function() {
	const _events = {};
	return {
		subscribe: (event, handler) => {
			let idx = 0;
			if (!_events[event]) {
				_events[event] = { [idx]: handler };
				return {
					unsubscribe: () => delete _events[event][idx],
				};
			}
			idx = Object.keys(_events[event]).length;
			_events[event][idx] = handler;
			return {
				unsubscribe: () => delete _events[event][idx],
			};
		},
		emit: (event, payload) => {
			if (!_events[event]) {
				return;
			}
			for (const evKey of Object.keys(_events[event])) {
				const ev = _events[event][evKey];
				if (!ev || typeof ev !== "function") {
					continue;
				}
				ev(payload ?? (null));
			}
		},
	}
};
window.alpinestorehandler.eventHandler = (function() {
	let _config = null;
	let _source = null;
	const _sourceCanReconnect = true;
	const _sourceMessage = new window.alpinestorehandler.eventEmitter();
	const _readyConnection = new window.alpinestorehandler.eventEmitter();

	function newMessage(event) {
		const message = JSON.parse(event.data);
		_sourceMessage.emit(message.type, message.payload);
	}

	function sourceError(event) {
		if (
			event?.target?.readyState === EventSource.CLOSED &&
			_sourceCanReconnect &&
			_config
		) {
			setTimeout(
				() =>
					_config ? open(_config) : sourceError(event),
				_config.reconnectTimeout
			);
		}
	}

	function sourceOpen(event) {
		_readyConnection.emit("ready");
	}

	function getClientId(key) {
		let clientId = localStorage.getItem(key);
		if (!clientId) {
			if (!crypto || typeof crypto.randomUUID !== 'function') {
				throw new Error('Crypto API not supported');
			}
			clientId = crypto.randomUUID();
			localStorage.setItem(key, clientId);
		}
		return clientId;
	}

	return {
		open: (config) => {
			if (!config.reconnectTimeout) {
				config.eventUrl = "/events";
				config.actionUrl = "/action";
				config.clientIdHeaderKey = "clientId";
				config.reconnectTimeout = 5000;
			}
			_config = config;
			if (_source) {
				_sourceCanReconnect = false;
				_source.close();
				_sourceCanReconnect = true;
				_source = null;
			}
			const pre = config.eventUrl.endsWith("/")
			? config.eventUrl.substring(0, config.eventUrl.length - 1)
			: config.eventUrl;
			const eventUrl = pre + '?clientId=' + getClientId(config.clientIdHeaderKey);
			_source = new EventSource(eventUrl);
			_source.onmessage = (event) =>
				newMessage(event);
			_source.onerror = (event) => sourceError(event);
			_source.onopen = (event) => sourceOpen(event);
		},
		subscribe: (event, handler) => {
			return _sourceMessage.subscribe(event, handler);
		},
		sendAction: async (message) => {
			if (!_config) {
				throw new Error("no config found");
			}
			return await fetch(_config.actionUrl, {
				mode: "cors",
				method: "POST",
				headers: {
					"Content-Type": "application/json",
					[_config.clientIdHeaderKey]: getClientId(
						_config.clientIdHeaderKey
					),
				},
				body: JSON.stringify(message),
			}).then((resp) => resp.json());
		}
	};
})();
`

func getJsScript() string {
	return src
}

func getAppScript(config Config) string {
	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(fmt.Sprintf(`window.alpinestorehandler.eventHandler.open({
		actionUrl: '%s',
		eventUrl: '%s',
		clientIdHeaderKey: '%s',
		reconnectTimeout: %v,
	});
	`, config.ActionUrl, config.EventUrl, config.ClientIDHeaderKey, config.SocketReconnectInterval))
	buf.WriteString("document.addEventListener('alpine:init', () => {")
	for _, h := range config.Handlers {
		writeStore(buf, h.GetName(), h.GetDefaultState(), h.GetActionType())
	}
	buf.WriteString("});")
	return buf.String()
}

func HeadScripts() string {
	return `
	<script src="//unpkg.com/alpinejs" defer></script>
	<script src="/alpinestorehandler_lib.js"></script>
	<script src="/alpinestorehandler_app.js"></script>
	`
}

func writeStore(buf *bytes.Buffer, name, defaultState, actionType string) {
	buf.WriteString(fmt.Sprintf(`
			Alpine.store('%[1]s', {
				state: %[2]v,
				emit(payload) {
					window.alpinestorehandler.eventHandler.sendAction({type:'%[3]s', payload});
				},
				update(state) {
					const keys = Object.keys(state);
					for(let i = 0; i < keys.length; i++) {
						this.state[keys[i]] = state[keys[i]];
					}
				}
			});
			window.alpinestorehandler.eventHandler.subscribe('[%[1]s] update', (payload) => {
				Alpine.store('%[1]s').update(payload);
			});
		`, name, defaultState, actionType))
}
