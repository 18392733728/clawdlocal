class ClawdLocalDashboard {
    constructor() {
        this.ws = null;
        this.init();
    }

    init() {
        this.setupWebSocket();
        this.setupEventListeners();
        this.updateStatus();
    }

    setupWebSocket() {
        const wsUrl = `ws://${window.location.host}/ws`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.showMessage('已连接到ClawdLocal服务', 'success');
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleWebSocketMessage(data);
        };

        this.ws.onclose = () => {
            console.log('WebSocket disconnected');
            this.showMessage('与ClawdLocal服务断开连接', 'error');
            // 尝试重连
            setTimeout(() => this.setupWebSocket(), 5000);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    handleWebSocketMessage(data) {
        switch(data.type) {
            case 'status_update':
                this.updateStatusDisplay(data.payload);
                break;
            case 'event':
                this.addEventToList(data.payload);
                break;
            case 'memory_update':
                this.updateMemoryDisplay(data.payload);
                break;
            case 'message_response':
                this.addChatMessage(data.payload, 'response');
                break;
            default:
                console.log('Unknown message type:', data.type);
        }
    }

    setupEventListeners() {
        const sendButton = document.getElementById('send-button');
        const messageInput = document.getElementById('message-input');

        sendButton.addEventListener('click', () => this.sendMessage());
        messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.sendMessage();
            }
        });
    }

    sendMessage() {
        const input = document.getElementById('message-input');
        const message = input.value.trim();
        
        if (!message) return;

        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const messageData = {
                type: 'message',
                payload: {
                    content: message,
                    timestamp: new Date().toISOString()
                }
            };
            
            this.ws.send(JSON.stringify(messageData));
            this.addChatMessage({content: message}, 'user');
            input.value = '';
        } else {
            this.showMessage('服务未连接，请稍后重试', 'error');
        }
    }

    addChatMessage(message, role) {
        const chatMessages = document.getElementById('chat-messages');
        const messageDiv = document.createElement('div');
        messageDiv.className = `chat-message ${role}`;
        messageDiv.textContent = message.content;
        chatMessages.appendChild(messageDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    updateStatus() {
        fetch('/api/status')
            .then(response => response.json())
            .then(data => this.updateStatusDisplay(data))
            .catch(error => {
                console.error('Failed to fetch status:', error);
                this.showMessage('无法获取系统状态', 'error');
            });
    }

    updateStatusDisplay(status) {
        const statusInfo = document.getElementById('status-info');
        statusInfo.innerHTML = `
            <p><strong>运行状态:</strong> ${status.running ? '运行中' : '已停止'}</p>
            <p><strong>版本:</strong> ${status.version}</p>
            <p><strong>工作空间:</strong> ${status.workspace}</p>
            <p><strong>事件队列长度:</strong> ${status.queue_length}</p>
            <p><strong>注册处理器:</strong> ${status.handlers_count}</p>
            <p><strong>内存使用:</strong> ${status.memory_usage}</p>
        `;
    }

    addEventToList(event) {
        const eventsList = document.getElementById('events-list');
        const eventDiv = document.createElement('div');
        eventDiv.className = 'event-item';
        eventDiv.innerHTML = `
            <strong>[${event.type}]</strong> ${new Date(event.timestamp).toLocaleString()}<br>
            <small>${JSON.stringify(event.data)}</small>
        `;
        eventsList.insertBefore(eventDiv, eventsList.firstChild);
        
        // 限制显示的事件数量
        if (eventsList.children.length > 20) {
            eventsList.removeChild(eventsList.lastChild);
        }
    }

    updateMemoryDisplay(memory) {
        const memoryInfo = document.getElementById('memory-info');
        memoryInfo.innerHTML = `
            <p><strong>短期记忆条目:</strong> ${memory.short_term_count}</p>
            <p><strong>长期记忆条目:</strong> ${memory.long_term_count}</p>
            <p><strong>最后更新:</strong> ${new Date(memory.last_updated).toLocaleString()}</p>
        `;
    }

    showMessage(text, type = 'info') {
        // 创建临时通知
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = text;
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// 初始化仪表板
document.addEventListener('DOMContentLoaded', () => {
    new ClawdLocalDashboard();
});