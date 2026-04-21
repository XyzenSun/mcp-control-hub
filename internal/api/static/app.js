class MCPGatewayUI {
    constructor() {
        this.apiKey = localStorage.getItem('mcp_api_key') || '';
        // Use current origin as base URL (same server)
        // For cases where API is on different origin, set window.MCP_API_BASE_URL in HTML
        this.baseURL = window.location.origin;
        this.currentView = 'servers';
        this.lang = localStorage.getItem('mcp_lang') || 'zh';
        this.data = {
            servers: [],
            tools: [],
            namespaces: [],
            apikeys: [],
            configs: []
        };

        this.i18n = {
            zh: {
                connected: '已连接',
                disconnected: '未连接',
                enabled: '已启用',
                disabled: '已禁用',
                sync: '同步',
                enable: '启用',
                disable: '禁用',
                edit: '编辑',
                delete: '删除',
                cancel: '取消',
                create: '创建',
                save: '保存',
                done: '完成',
                revoke: '撤销',
                never: '从未',
                noDescription: '无描述',
                noServersConfigured: '未配置服务器',
                addFirstServer: '添加您的第一个上游 MCP 服务器以开始使用。',
                noNamespacesCreated: '未创建命名空间',
                createNamespaces: '创建命名空间以组织和隔离工具。',
                noToolsFound: '未找到工具',
                noApiKeysGenerated: '未生成 API 密钥',
                serverCreated: '服务器已创建',
                serverUpdated: '服务器已更新',
                serverSynced: '服务器已同步',
                serverEnabled: '服务器已启用',
                serverDisabled: '服务器已禁用',
                serverDeleted: '服务器已删除',
                toolsRefreshed: '工具已刷新',
                toolEnabled: '工具已启用',
                toolDisabled: '工具已禁用',
                namespaceCreated: '命名空间已创建',
                namespaceDeleted: '命名空间已删除',
                apiKeyRevoked: 'API 密钥已撤销',
                deleteServerConfirm: '删除此服务器？',
                deleteNamespaceConfirm: '删除此命名空间？',
                revokeApiKeyConfirm: '撤销此 API 密钥？',
                configuration: '配置',
                gatewayUrl: 'Gateway URL',
                apiKey: 'API 密钥',
                addServer: '添加服务器',
                editServer: '编辑服务器',
                name: '名称',
                protocol: '协议',
                description: '描述',
                config: '配置 (JSON)',
                configHint: '提供协议特定的 JSON 配置。',
                addNamespace: '添加命名空间',
                manageTools: '管理工具',
                generateApiKey: '生成 API 密钥',
                keyName: '密钥名称',
                apiKeyGenerated: 'API 密钥已生成',
                saveKeyNow: '立即保存此密钥',
                keyShownOnce: '此密钥仅显示一次。',
                selectTools: '选择工具',
                toolsInNamespace: '命名空间中的工具',
                configSaved: '配置已保存',
                configReset: '配置已重置',
                configSaveFailed: '保存失败',
                saveAll: '保存全部',
                timeoutSettings: '超时设置',
                intervalSettings: '间隔设置',
                seconds: '秒',
                resetConfirm: '确定要重置所有配置为默认值吗？'
            },
            en: {
                connected: 'Connected',
                disconnected: 'Disconnected',
                enabled: 'Enabled',
                disabled: 'Disabled',
                sync: 'Sync',
                enable: 'Enable',
                disable: 'Disable',
                edit: 'Edit',
                delete: 'Delete',
                cancel: 'Cancel',
                create: 'Create',
                save: 'Save',
                done: 'Done',
                revoke: 'Revoke',
                never: 'Never',
                noDescription: 'No description',
                noServersConfigured: 'No servers configured',
                addFirstServer: 'Add your first upstream MCP server to get started.',
                noNamespacesCreated: 'No namespaces created',
                createNamespaces: 'Create namespaces to organize and isolate tools.',
                noToolsFound: 'No tools found',
                noApiKeysGenerated: 'No API keys generated',
                serverCreated: 'Server created',
                serverUpdated: 'Server updated',
                serverSynced: 'Server synced',
                serverEnabled: 'Server enabled',
                serverDisabled: 'Server disabled',
                serverDeleted: 'Server deleted',
                toolsRefreshed: 'Tools refreshed',
                toolEnabled: 'Tool enabled',
                toolDisabled: 'Tool disabled',
                namespaceCreated: 'Namespace created',
                namespaceDeleted: 'Namespace deleted',
                apiKeyRevoked: 'API key revoked',
                deleteServerConfirm: 'Delete this server?',
                deleteNamespaceConfirm: 'Delete this namespace?',
                revokeApiKeyConfirm: 'Revoke this API key?',
                configuration: 'Configuration',
                gatewayUrl: 'Gateway URL',
                apiKey: 'API Key',
                addServer: 'Add Server',
                editServer: 'Edit Server',
                name: 'Name',
                protocol: 'Protocol',
                description: 'Description',
                config: 'Config (JSON)',
                configHint: 'Provide protocol-specific configuration as JSON.',
                addNamespace: 'Add Namespace',
                manageTools: 'Manage Tools',
                generateApiKey: 'Generate API Key',
                keyName: 'Key Name',
                apiKeyGenerated: 'API Key Generated',
                saveKeyNow: 'Save this key now',
                keyShownOnce: 'This key will only be shown once.',
                selectTools: 'Select Tools',
                toolsInNamespace: 'Tools in Namespace',
                configSaved: 'Config saved',
                configReset: 'Config reset',
                configSaveFailed: 'Save failed',
                saveAll: 'Save All',
                timeoutSettings: 'Timeout Settings',
                intervalSettings: 'Interval Settings',
                seconds: 'seconds',
                resetConfirm: 'Reset all configurations to default values?'
            }
        };

        this.init();
    }

    init() {
        this.checkAuth();
        this.bindEvents();
        this.applyLanguage();
        this.checkConnection();
        this.loadData();
    }

    checkAuth() {
        if (!this.apiKey) {
            window.location.href = '/login.html';
            return;
        }
    }

    logout() {
        localStorage.removeItem('mcp_api_key');
        window.location.href = '/login.html';
    }

    bindEvents() {
        // Navigation
        document.querySelectorAll('.nav-item').forEach(btn => {
            btn.addEventListener('click', () => this.switchView(btn.dataset.view));
        });

        // Language toggle
        document.getElementById('langToggle').addEventListener('click', () => this.toggleLanguage());

        // Logout
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());

        // Modal
        document.getElementById('modalClose').addEventListener('click', () => this.closeModal());
        document.getElementById('modal').addEventListener('click', (e) => {
            if (e.target.id === 'modal') this.closeModal();
        });

        // Action buttons
        document.getElementById('addServerBtn').addEventListener('click', () => this.showServerModal());
        document.getElementById('refreshToolsBtn').addEventListener('click', () => this.refreshTools());
        document.getElementById('addNamespaceBtn').addEventListener('click', () => this.showNamespaceModal());
        document.getElementById('generateKeyBtn').addEventListener('click', () => this.showAPIKeyModal());

        // Filters
        document.getElementById('serverFilter').addEventListener('change', () => this.renderTools());
        document.getElementById('enabledFilter').addEventListener('change', () => this.renderTools());

        // Config
        document.getElementById('resetConfigBtn').addEventListener('click', () => this.resetConfig());
    }

    toggleLanguage() {
        this.lang = this.lang === 'zh' ? 'en' : 'zh';
        localStorage.setItem('mcp_lang', this.lang);
        document.getElementById('langToggle').textContent = this.lang === 'zh' ? 'EN' : '中文';
        this.applyLanguage();
        this.render();
    }

    applyLanguage() {
        document.querySelectorAll('[data-zh][data-en]').forEach(el => {
            el.textContent = this.lang === 'zh' ? el.dataset.zh : el.dataset.en;
        });
        document.getElementById('langToggle').textContent = this.lang === 'zh' ? 'EN' : '中文';
    }

    t(key) {
        return this.i18n[this.lang][key] || key;
    }

    async request(endpoint, options = {}) {
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (this.apiKey) {
            headers['X-API-Key'] = this.apiKey;
        }

        try {
            const response = await fetch(`${this.baseURL}${endpoint}`, {
                ...options,
                headers
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }

            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return await response.json();
            }
            return await response.text();
        } catch (error) {
            console.error('Request failed:', error);
            this.showToast(`Request failed: ${error.message}`);
            throw error;
        }
    }

    async checkConnection() {
        try {
            const response = await fetch(`${this.baseURL}/health`);
            if (!response.ok) {
                throw new Error('Health check failed');
            }
            this.updateStatus(true);
        } catch {
            this.updateStatus(false);
            // If connection fails, redirect to login
            this.logout();
        }
    }

    updateStatus(online) {
        const indicator = document.getElementById('statusIndicator');
        const text = document.getElementById('statusText');

        indicator.classList.toggle('online', online);
        text.textContent = online ? this.t('connected') : this.t('disconnected');
    }

    async loadData() {
        try {
            const [servers, tools, namespaces, apikeys, configs] = await Promise.all([
                this.request('/api/v1/servers').catch(() => []),
                this.request('/api/v1/tools').catch(() => []),
                this.request('/api/v1/namespaces').catch(() => []),
                this.request('/api/v1/apikeys').catch(() => []),
                this.request('/api/v1/config').catch(() => [])
            ]);

            this.data.servers = servers;
            this.data.tools = tools;
            this.data.namespaces = namespaces;
            this.data.apikeys = apikeys;
            this.data.configs = configs;

            this.render();
            this.populateServerFilter();
        } catch (error) {
            console.error('Failed to load data:', error);
            // If data loading fails due to auth, redirect to login
            if (error.message.includes('401') || error.message.includes('403')) {
                this.logout();
            }
        }
    }

    render() {
        this.renderServers();
        this.renderTools();
        this.renderNamespaces();
        this.renderAPIKeys();
        this.renderConfig();
    }

    switchView(view) {
        this.currentView = view;

        document.querySelectorAll('.nav-item').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.view === view);
        });

        document.querySelectorAll('.view').forEach(section => {
            section.classList.toggle('active', section.id === `${view}View`);
        });
    }

    renderServers() {
        const container = document.getElementById('serversList');

        if (!this.data.servers.length) {
            container.innerHTML = this.emptyState(this.t('noServersConfigured'), this.t('addFirstServer'));
            return;
        }

        container.innerHTML = this.data.servers.map(server => `
            <div class="card">
                <div class="card-header">
                    <div>
                        <h3 class="card-title">${server.name}</h3>
                        <p class="card-subtitle">${server.protocol}</p>
                    </div>
                    <span class="card-badge ${server.enabled ? 'enabled' : ''}">${server.enabled ? this.t('enabled') : this.t('disabled')}</span>
                </div>
                <div class="card-body">
                    <p class="card-description">${server.description || this.t('noDescription')}</p>
                    <div class="card-meta">
                        <div class="card-meta-item">
                            <span class="card-meta-label">ID</span>
                            <span class="card-meta-value">${server.id}</span>
                        </div>
                        <div class="card-meta-item">
                            <span class="card-meta-label">${this.t('protocol')}</span>
                            <span class="card-meta-value">${server.protocol}</span>
                        </div>
                    </div>
                </div>
                <div class="card-actions">
                    <button class="btn btn-small" onclick="app.syncServer(${server.id})">${this.t('sync')}</button>
                    <button class="btn btn-small" onclick="app.editServer(${server.id})">${this.t('edit')}</button>
                    <button class="btn btn-small" onclick="app.toggleServer(${server.id}, ${!server.enabled})">
                        ${server.enabled ? this.t('disable') : this.t('enable')}
                    </button>
                    <button class="btn btn-small btn-danger" onclick="app.deleteServer(${server.id})">${this.t('delete')}</button>
                </div>
            </div>
        `).join('');
    }

    renderTools() {
        const tbody = document.getElementById('toolsTableBody');
        const serverFilter = document.getElementById('serverFilter').value;
        const enabledFilter = document.getElementById('enabledFilter').value;

        let tools = [...this.data.tools];

        if (serverFilter) {
            tools = tools.filter(tool => String(tool.server_id) === serverFilter);
        }

        if (enabledFilter) {
            const enabled = enabledFilter === 'true';
            tools = tools.filter(tool => tool.enabled === enabled);
        }

        if (!tools.length) {
            tbody.innerHTML = `<tr><td colspan="5" class="empty-state-text">${this.t('noToolsFound')}</td></tr>`;
            return;
        }

        tbody.innerHTML = tools.map(tool => {
            const server = this.data.servers.find(s => s.id === tool.server_id);
            const useOverride = tool.use_override && tool.override_description;
            const displayDesc = useOverride ? tool.override_description : tool.original_description;
            const overrideBadge = useOverride ? `<span class="card-badge" style="margin-left: var(--spacing-xs); font-size: var(--font-size-xs);">${this.lang === 'zh' ? '自定义' : 'Custom'}</span>` : '';
            return `
                <tr>
                    <td>${tool.name}</td>
                    <td>${server ? server.name : '-'}</td>
                    <td>${displayDesc || '-'}${overrideBadge}</td>
                    <td>
                        <span class="card-badge ${tool.enabled ? 'enabled' : ''}">
                            ${tool.enabled ? this.t('enabled') : this.t('disabled')}
                        </span>
                    </td>
                    <td>
                        <div class="table-actions">
                            <button class="btn btn-small" onclick="app.editTool(${tool.id})">${this.t('edit')}</button>
                            <button class="btn btn-small" onclick="app.toggleTool(${tool.id}, ${!tool.enabled})">
                                ${tool.enabled ? this.t('disable') : this.t('enable')}
                            </button>
                        </div>
                    </td>
                </tr>
            `;
        }).join('');
    }

    renderNamespaces() {
        const container = document.getElementById('namespacesList');

        if (!this.data.namespaces.length) {
            container.innerHTML = this.emptyState(this.t('noNamespacesCreated'), this.t('createNamespaces'));
            return;
        }

        container.innerHTML = this.data.namespaces.map(namespace => `
            <div class="card">
                <div class="card-header">
                    <div>
                        <h3 class="card-title">${namespace.name}</h3>
                        <p class="card-subtitle">${this.lang === 'zh' ? '命名空间' : 'Namespace'}</p>
                    </div>
                </div>
                <div class="card-body">
                    <p class="card-description">${namespace.description || this.t('noDescription')}</p>
                    <div class="card-meta">
                        <div class="card-meta-item">
                            <span class="card-meta-label">ID</span>
                            <span class="card-meta-value">${namespace.id}</span>
                        </div>
                    </div>
                </div>
                <div class="card-actions">
                    <button class="btn btn-small" onclick="app.manageNamespaceTools('${namespace.id}')">${this.t('manageTools')}</button>
                    <button class="btn btn-small btn-danger" onclick="app.deleteNamespace('${namespace.id}')">${this.t('delete')}</button>
                </div>
            </div>
        `).join('');
    }

    renderAPIKeys() {
        const tbody = document.getElementById('apikeysTableBody');

        if (!this.data.apikeys.length) {
            tbody.innerHTML = `<tr><td colspan="4" class="empty-state-text">${this.t('noApiKeysGenerated')}</td></tr>`;
            return;
        }

        tbody.innerHTML = this.data.apikeys.map(key => `
            <tr>
                <td>${key.name}</td>
                <td>${this.formatDate(key.created_at)}</td>
                <td>${key.last_used ? this.formatDate(key.last_used) : this.t('never')}</td>
                <td>
                    <button class="btn btn-small btn-danger" onclick="app.deleteAPIKey(${key.id})">${this.t('revoke')}</button>
                </td>
            </tr>
        `).join('');
    }

    populateServerFilter() {
        const select = document.getElementById('serverFilter');
        const allServersText = this.lang === 'zh' ? '所有服务器' : 'All Servers';
        select.innerHTML = `<option value="">${allServersText}</option>` +
            this.data.servers.map(server => `<option value="${server.id}">${server.name}</option>`).join('');
    }

    // Remove showConfigModal as we now use login page

    showServerModal() {
        this.showModal(this.t('addServer'), `
            <form id="serverForm">
                <div class="form-group">
                    <label class="form-label">${this.t('name')}</label>
                    <input class="form-input" name="name" required>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('protocol')}</label>
                    <select class="form-select" name="protocol" required>
                        <option value="streamable">Streamable HTTP</option>
                        <option value="sse">SSE</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('description')}</label>
                    <textarea class="form-textarea" name="description"></textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('config')}</label>
                    <textarea class="form-textarea" name="config" placeholder='{"url":"https://example.com/mcp"}' required></textarea>
                    <div class="form-hint">${this.t('configHint')}</div>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('create')}</button>
                </div>
            </form>
        `);

        document.getElementById('serverForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            try {
                await this.request('/api/v1/servers', {
                    method: 'POST',
                    body: JSON.stringify({
                        name: form.get('name'),
                        protocol: form.get('protocol'),
                        description: form.get('description'),
                        config: form.get('config'),
                        enabled: true
                    })
                });
                this.closeModal();
                this.showToast(this.t('serverCreated'));
                this.loadData();
            } catch {}
        });
    }

    editServer(id) {
        const server = this.data.servers.find(s => s.id === id);
        if (!server) return;

        this.showModal(this.t('editServer'), `
            <form id="editServerForm">
                <div class="form-group">
                    <label class="form-label">${this.t('name')}</label>
                    <input class="form-input" name="name" value="${server.name}" required>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('protocol')}</label>
                    <select class="form-select" name="protocol" required>
                        <option value="streamable" ${server.protocol === 'streamable' ? 'selected' : ''}>Streamable HTTP</option>
                        <option value="sse" ${server.protocol === 'sse' ? 'selected' : ''}>SSE</option>
                    </select>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('description')}</label>
                    <textarea class="form-textarea" name="description">${server.description || ''}</textarea>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('config')}</label>
                    <textarea class="form-textarea" name="config" required>${server.config}</textarea>
                    <div class="form-hint">${this.t('configHint')}</div>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('save')}</button>
                </div>
            </form>
        `);

        document.getElementById('editServerForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            try {
                await this.request(`/api/v1/servers/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify({
                        name: form.get('name'),
                        protocol: form.get('protocol'),
                        description: form.get('description'),
                        config: form.get('config'),
                        enabled: server.enabled
                    })
                });
                this.closeModal();
                this.showToast(this.t('serverUpdated'));
                this.loadData();
            } catch {}
        });
    }

    showNamespaceModal() {
        this.showModal(this.t('addNamespace'), `
            <form id="namespaceForm">
                <div class="form-group">
                    <label class="form-label">ID</label>
                    <input class="form-input" name="id" pattern="[a-zA-Z0-9]+" maxlength="50" required>
                    <div class="form-hint">${this.lang === 'zh' ? '仅限字母和数字，用于 URL' : 'Alphanumeric only, used in URLs'}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('name')}</label>
                    <input class="form-input" name="name" maxlength="16" required>
                    <div class="form-hint">${this.lang === 'zh' ? '显示名称，1-16 个字符' : 'Display name, 1-16 characters'}</div>
                </div>
                <div class="form-group">
                    <label class="form-label">${this.t('description')}</label>
                    <textarea class="form-textarea" name="description"></textarea>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('create')}</button>
                </div>
            </form>
        `);

        document.getElementById('namespaceForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            try {
                await this.request('/api/v1/namespaces', {
                    method: 'POST',
                    body: JSON.stringify({
                        id: form.get('id'),
                        name: form.get('name'),
                        description: form.get('description')
                    })
                });
                this.closeModal();
                this.showToast(this.t('namespaceCreated'));
                this.loadData();
            } catch {}
        });
    }

    async manageNamespaceTools(namespaceId) {
        const namespace = this.data.namespaces.find(ns => ns.id === namespaceId);
        if (!namespace) return;

        // Get tools in this namespace
        let namespaceTools = [];
        try {
            namespaceTools = await this.request(`/api/v1/namespaces/${namespaceId}/tools`);
        } catch {
            namespaceTools = [];
        }

        const namespaceToolIds = new Set(namespaceTools.map(t => t.id));

        // Group tools by server
        const toolsByServer = {};
        this.data.tools.forEach(tool => {
            const server = this.data.servers.find(s => s.id === tool.server_id);
            const serverName = server ? server.name : 'Unknown';
            if (!toolsByServer[serverName]) {
                toolsByServer[serverName] = [];
            }
            toolsByServer[serverName].push(tool);
        });

        const toolsHTML = Object.entries(toolsByServer).map(([serverName, tools]) => `
            <div style="margin-bottom: var(--spacing-lg);">
                <h4 style="font-size: var(--font-size-sm); font-weight: var(--font-weight-semibold); margin-bottom: var(--spacing-sm); color: var(--color-text-secondary);">${serverName}</h4>
                ${tools.map(tool => `
                    <label style="display: flex; align-items: center; padding: var(--spacing-sm); cursor: pointer; border-radius: var(--radius-sm); transition: background var(--transition-base);" onmouseover="this.style.background='var(--color-bg-secondary)'" onmouseout="this.style.background='transparent'">
                        <input type="checkbox" name="tools" value="${tool.id}" ${namespaceToolIds.has(tool.id) ? 'checked' : ''} style="margin-right: var(--spacing-sm);">
                        <span style="flex: 1; font-size: var(--font-size-sm);">${tool.name}</span>
                        ${!tool.enabled ? '<span style="font-size: var(--font-size-xs); color: var(--color-text-tertiary);">(disabled)</span>' : ''}
                    </label>
                `).join('')}
            </div>
        `).join('');

        this.showModal(`${this.t('manageTools')} - ${namespace.name}`, `
            <form id="manageToolsForm">
                <div style="max-height: 400px; overflow-y: auto; margin-bottom: var(--spacing-lg);">
                    ${toolsHTML || '<p style="text-align: center; color: var(--color-text-secondary);">' + this.t('noToolsFound') + '</p>'}
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('save')}</button>
                </div>
            </form>
        `);

        document.getElementById('manageToolsForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            const selectedToolIds = form.getAll('tools').map(id => parseInt(id));

            try {
                // Remove all tools first, then add selected ones
                for (const tool of namespaceTools) {
                    await this.request(`/api/v1/namespaces/${namespaceId}/tools/${tool.id}`, {
                        method: 'DELETE'
                    });
                }

                // Add selected tools
                for (const toolId of selectedToolIds) {
                    await this.request(`/api/v1/namespaces/${namespaceId}/tools`, {
                        method: 'POST',
                        body: JSON.stringify({ tool_id: toolId })
                    });
                }

                this.closeModal();
                this.showToast(this.t('namespaceCreated')); // Reuse translation
                this.loadData();
            } catch (error) {
                console.error('Failed to update namespace tools:', error);
            }
        });
    }

    showAPIKeyModal() {
        this.showModal(this.t('generateApiKey'), `
            <form id="apikeyForm">
                <div class="form-group">
                    <label class="form-label">${this.t('keyName')}</label>
                    <input class="form-input" name="name" required>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('create')}</button>
                </div>
            </form>
        `);

        document.getElementById('apikeyForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            try {
                const result = await this.request('/api/v1/apikeys', {
                    method: 'POST',
                    body: JSON.stringify({ name: form.get('name') })
                });
                this.closeModal();
                this.showModal(this.t('apiKeyGenerated'), `
                    <div class="form-group">
                        <label class="form-label">${this.t('saveKeyNow')}</label>
                        <textarea class="form-textarea" readonly>${result.key || result.api_key || 'Key created successfully'}</textarea>
                        <div class="form-hint">${this.t('keyShownOnce')}</div>
                    </div>
                    <div class="form-actions">
                        <button class="btn btn-primary" onclick="app.closeModal()">${this.t('done')}</button>
                    </div>
                `);
                this.loadData();
            } catch {}
        });
    }

    async syncServer(id) {
        try {
            await this.request(`/api/v1/servers/${id}/sync`, { method: 'POST' });
            this.showToast(this.t('serverSynced'));
            this.loadData();
        } catch {}
    }

    async toggleServer(id, enabled) {
        try {
            await this.request(`/api/v1/servers/${id}/${enabled ? 'enable' : 'disable'}`, { method: 'POST' });
            this.showToast(enabled ? this.t('serverEnabled') : this.t('serverDisabled'));
            this.loadData();
        } catch {}
    }

    async deleteServer(id) {
        if (!confirm(this.t('deleteServerConfirm'))) return;
        try {
            await this.request(`/api/v1/servers/${id}`, { method: 'DELETE' });
            this.showToast(this.t('serverDeleted'));
            this.loadData();
        } catch {}
    }

    async refreshTools() {
        try {
            await this.request('/api/v1/tools/refresh', { method: 'POST' });
            this.showToast(this.t('toolsRefreshed'));
            this.loadData();
        } catch {}
    }

    async toggleTool(id, enabled) {
        try {
            await this.request(`/api/v1/tools/${id}/${enabled ? 'enable' : 'disable'}`, { method: 'POST' });
            this.showToast(enabled ? this.t('toolEnabled') : this.t('toolDisabled'));
            this.loadData();
        } catch {}
    }

    editTool(id) {
        const tool = this.data.tools.find(t => t.id === id);
        if (!tool) return;

        const server = this.data.servers.find(s => s.id === tool.server_id);
        const useOverride = tool.use_override || false;

        this.showModal(this.t('edit') + ' - ' + tool.name, `
            <form id="editToolForm">
                <div class="form-group">
                    <label class="form-label">${this.t('name')}</label>
                    <input class="form-input" value="${tool.name}" disabled>
                    <div class="form-hint">工具名称由上游服务器定义，不可修改</div>
                </div>
                <div class="form-group">
                    <label class="form-label">服务器</label>
                    <input class="form-input" value="${server ? server.name : '-'}" disabled>
                </div>
                <div class="form-group">
                    <label class="form-label">原始描述</label>
                    <textarea class="form-textarea" disabled>${tool.original_description || '无'}</textarea>
                </div>
                <div class="form-group">
                    <label class="form-label" style="display: flex; align-items: center; gap: var(--spacing-sm);">
                        <input type="checkbox" name="use_override" ${useOverride ? 'checked' : ''} onchange="document.getElementById('overrideDescField').style.display = this.checked ? 'block' : 'none'">
                        <span>${this.lang === 'zh' ? '使用自定义描述' : 'Use custom description'}</span>
                    </label>
                    <div class="form-hint">${this.lang === 'zh' ? '启用后将使用自定义描述替代原始描述' : 'When enabled, custom description overrides original'}</div>
                </div>
                <div class="form-group" id="overrideDescField" style="display: ${useOverride ? 'block' : 'none'};">
                    <label class="form-label">${this.lang === 'zh' ? '自定义描述' : 'Custom Description'}</label>
                    <textarea class="form-textarea" name="override_description">${tool.override_description || ''}</textarea>
                </div>
                <div class="form-actions">
                    <button type="button" class="btn" onclick="app.closeModal()">${this.t('cancel')}</button>
                    <button type="submit" class="btn btn-primary">${this.t('save')}</button>
                </div>
            </form>
        `);

        document.getElementById('editToolForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const form = new FormData(e.target);
            const useOverride = form.get('use_override') === 'on';
            const overrideDescription = form.get('override_description')?.trim() || '';

            try {
                await this.request(`/api/v1/tools/${id}`, {
                    method: 'PUT',
                    body: JSON.stringify({
                        use_override: useOverride,
                        override_description: overrideDescription || null
                    })
                });
                this.closeModal();
                this.showToast(this.t('serverUpdated')); // Reuse translation
                this.loadData();
            } catch (error) {
                console.error('Failed to update tool:', error);
            }
        });
    }

    async deleteNamespace(id) {
        if (!confirm(this.t('deleteNamespaceConfirm'))) return;
        try {
            await this.request(`/api/v1/namespaces/${id}`, { method: 'DELETE' });
            this.showToast(this.t('namespaceDeleted'));
            this.loadData();
        } catch {}
    }

    async deleteAPIKey(id) {
        if (!confirm(this.t('revokeApiKeyConfirm'))) return;
        try {
            await this.request(`/api/v1/apikeys/${id}`, { method: 'DELETE' });
            this.showToast(this.t('apiKeyRevoked'));
            this.loadData();
        } catch (error) {
            console.error('Failed to delete API key:', error);
        }
    }

    showModal(title, content) {
        document.getElementById('modalTitle').textContent = title;
        document.getElementById('modalBody').innerHTML = content;
        document.getElementById('modal').classList.add('active');
    }

    closeModal() {
        document.getElementById('modal').classList.remove('active');
    }

    showToast(message) {
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.classList.add('active');
        clearTimeout(this.toastTimer);
        this.toastTimer = setTimeout(() => {
            toast.classList.remove('active');
        }, 3000);
    }

    emptyState(title, text) {
        return `
            <div class="empty-state">
                <h3 class="empty-state-title">${title}</h3>
                <p class="empty-state-text">${text}</p>
            </div>
        `;
    }

    formatDate(dateString) {
        return new Date(dateString).toLocaleDateString('zh-CN', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    }

    renderConfig() {
        const container = document.getElementById('configContainer');

        if (!this.data.configs || !this.data.configs.length) {
            container.innerHTML = this.emptyState(
                this.lang === 'zh' ? '无配置数据' : 'No config data',
                this.lang === 'zh' ? '无法加载配置信息' : 'Could not load configuration'
            );
            return;
        }

        // Group configs by category
        const timeoutConfigs = this.data.configs.filter(c => c.category === 'timeout');
        const intervalConfigs = this.data.configs.filter(c => c.category === 'interval');

        const renderConfigGroup = (configs, title) => `
            <div class="config-group">
                <h3 class="config-group-title">${title}</h3>
                <div class="config-items">
                    ${configs.map(config => `
                        <div class="config-item">
                            <div class="config-item-info">
                                <label class="config-label" for="config_${config.key}">${this.translateConfigKey(config.key)}</label>
                                <div class="config-hint">${config.description}</div>
                            </div>
                            <div class="config-item-input">
                                <input type="number"
                                    id="config_${config.key}"
                                    class="form-input config-input"
                                    data-key="${config.key}"
                                    data-default="${config.default_value}"
                                    value="${config.value || config.default_value}"
                                    min="1"
                                    max="300">
                                <span class="config-unit">${this.t('seconds')}</span>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;

        container.innerHTML = `
            ${renderConfigGroup(timeoutConfigs, this.t('timeoutSettings'))}
            ${renderConfigGroup(intervalConfigs, this.t('intervalSettings'))}
            <div class="form-actions" style="margin-top: var(--spacing-xl);">
                <button class="btn btn-primary" onclick="app.saveAllConfig()">${this.t('saveAll')}</button>
            </div>
        `;
    }

    translateConfigKey(key) {
        const translations = {
            list_tools_timeout: this.lang === 'zh' ? 'ListTools 超时' : 'ListTools Timeout',
            call_tool_timeout: this.lang === 'zh' ? 'CallTool 超时' : 'CallTool Timeout',
            connect_timeout: this.lang === 'zh' ? '连接超时' : 'Connect Timeout',
            health_check_timeout: this.lang === 'zh' ? '健康检查超时' : 'Health Check Timeout',
            health_check_interval: this.lang === 'zh' ? '健康检查间隔' : 'Health Check Interval'
        };
        return translations[key] || key;
    }

    async saveAllConfig() {
        const inputs = document.querySelectorAll('.config-input');
        const updates = {};

        inputs.forEach(input => {
            const key = input.dataset.key;
            const value = parseInt(input.value, 10);
            if (key && !isNaN(value) && value > 0) {
                updates[key] = value.toString();
            }
        });

        try {
            await this.request('/api/v1/config', {
                method: 'PUT',
                body: JSON.stringify(updates)
            });
            this.showToast(this.t('configSaved'));
            this.loadData();
        } catch (error) {
            console.error('Failed to save config:', error);
            this.showToast(this.t('configSaveFailed'));
        }
    }

    async resetConfig() {
        if (!confirm(this.t('resetConfirm'))) return;

        try {
            // Reset all configs to their default values
            const updates = {};
            this.data.configs.forEach(config => {
                updates[config.key] = config.default_value;
            });

            await this.request('/api/v1/config', {
                method: 'PUT',
                body: JSON.stringify(updates)
            });
            this.showToast(this.t('configReset'));
            this.loadData();
        } catch (error) {
            console.error('Failed to reset config:', error);
            this.showToast(this.t('configSaveFailed'));
        }
    }
}

const app = new MCPGatewayUI();
