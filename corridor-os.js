// Corridor OS - Main Operating System Logic
class CorridorOS {
    constructor() {
        this.isBooted = false;
        this.isLocked = false;
        this.currentWorkspace = 1;
        this.openWindows = new Map();
        this.windowZIndex = 1000;
        this.notifications = [];
        this.systemTime = new Date();
        
        // System state
        this.systemInfo = {
            version: '1.0.0',
            buildNumber: '20241217',
            kernel: 'Corridor Quantum Kernel',
            architecture: 'x86_64',
            memory: '16GB',
            storage: '512GB NVMe',
            cpu: 'Quantum Processing Unit',
            gpu: 'Photonic Rendering Engine'
        };
        
        // User preferences
        this.preferences = {
            theme: 'dark',
            wallpaper: 'quantum-gradient',
            animations: true,
            notifications: true,
            autoLock: 300000, // 5 minutes
            language: 'en-US',
            timezone: 'UTC'
        };
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.startBootSequence();
        this.updateClock();
        setInterval(() => this.updateClock(), 1000);
    }
    
    setupEventListeners() {
        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => this.handleKeyboard(e));
        
        // Mouse events
        document.addEventListener('contextmenu', (e) => this.handleContextMenu(e));
        document.addEventListener('click', (e) => this.handleClick(e));
        
        // Window management
        document.addEventListener('mousedown', (e) => this.handleWindowInteraction(e));
        document.addEventListener('mousemove', (e) => this.handleMouseMove(e));
        document.addEventListener('mouseup', (e) => this.handleMouseUp(e));
        
        // Prevent default behaviors
        document.addEventListener('dragstart', (e) => e.preventDefault());
        document.addEventListener('drop', (e) => e.preventDefault());
        document.addEventListener('dragover', (e) => e.preventDefault());
    }
    
    async startBootSequence() {
        const bootSteps = [
            { message: 'Initializing quantum substrate...', duration: 800 },
            { message: 'Loading photonic drivers...', duration: 600 },
            { message: 'Calibrating optical pathways...', duration: 700 },
            { message: 'Starting memory mesh...', duration: 500 },
            { message: 'Initializing heliopass sensors...', duration: 400 },
            { message: 'Loading thermal equilibrium model...', duration: 600 },
            { message: 'Starting orchestrator services...', duration: 500 },
            { message: 'Mounting quantum filesystem...', duration: 400 },
            { message: 'Loading user environment...', duration: 300 },
            { message: 'Corridor OS ready!', duration: 500 }
        ];
        
        const progressFill = document.getElementById('boot-progress-fill');
        const statusElement = document.getElementById('boot-status');
        
        for (let i = 0; i < bootSteps.length; i++) {
            const step = bootSteps[i];
            statusElement.textContent = step.message;
            progressFill.style.width = `${((i + 1) / bootSteps.length) * 100}%`;
            
            await new Promise(resolve => setTimeout(resolve, step.duration));
        }
        
        // Boot complete
        await new Promise(resolve => setTimeout(resolve, 500));
        this.completeBootSequence();
    }
    
    completeBootSequence() {
        this.isBooted = true;
        const bootSplash = document.getElementById('boot-splash');
        const desktop = document.getElementById('desktop');
        
        if (bootSplash) bootSplash.style.display = 'none';
        if (desktop) desktop.style.display = 'block';
        
        // Show welcome notification
        this.showNotification('Welcome to Corridor OS', 'Hybrid Quantum-Photonic Operating System ready for use.');
        
        // Auto-lock timer
        this.resetAutoLockTimer();
    }
    
    updateClock() {
        this.systemTime = new Date();
        const timeString = this.systemTime.toLocaleTimeString('en-US', { 
            hour12: false, 
            hour: '2-digit', 
            minute: '2-digit' 
        });
        const dateString = this.systemTime.toLocaleDateString('en-US', { 
            weekday: 'long', 
            year: 'numeric', 
            month: 'long', 
            day: 'numeric' 
        });
        
        // Update desktop clock
        const desktopClock = document.getElementById('desktop-clock');
        if (desktopClock) {
            desktopClock.textContent = timeString;
        }
        
        // Update lock screen clock
        const lockClock = document.getElementById('lock-clock');
        const lockDate = document.getElementById('lock-date');
        if (lockClock && lockDate) {
            lockClock.textContent = timeString;
            lockDate.textContent = dateString;
        }
    }
    
    handleKeyboard(e) {
        // Global shortcuts
        if (e.altKey && e.key === 'F4') {
            e.preventDefault();
            this.closeActiveWindow();
        }
        
        if (e.metaKey || e.ctrlKey) {
            switch (e.key) {
                case ' ':
                    e.preventDefault();
                    this.showActivitiesOverview();
                    break;
                case 't':
                    e.preventDefault();
                    this.openApp('terminal');
                    break;
                case 'l':
                    e.preventDefault();
                    this.lockScreen();
                    break;
                case ',':
                    e.preventDefault();
                    this.openApp('settings');
                    break;
            }
        }
        
        // Function keys
        if (e.key.startsWith('F') && !e.altKey && !e.ctrlKey && !e.metaKey) {
            const fNumber = parseInt(e.key.substring(1));
            if (fNumber >= 1 && fNumber <= 12) {
                e.preventDefault();
                this.handleFunctionKey(fNumber);
            }
        }
        
        // Escape key
        if (e.key === 'Escape') {
            this.hideOverlays();
        }
    }
    
    handleFunctionKey(number) {
        const functionKeyActions = {
            1: () => this.showHelp(),
            2: () => this.openApp('files'),
            3: () => this.openApp('terminal'),
            4: () => this.openApp('text-editor'),
            5: () => this.refreshDesktop(),
            6: () => this.openApp('quantum-lab'),
            7: () => this.openApp('photonic-studio'),
            8: () => this.openApp('corridor-computer'),
            9: () => this.openApp('settings'),
            10: () => this.openApp('web-browser'),
            11: () => this.toggleFullscreen(),
            12: () => this.openApp('system-monitor')
        };
        
        const action = functionKeyActions[number];
        if (action) {
            action();
        }
    }
    
    handleContextMenu(e) {
        e.preventDefault();
        this.showContextMenu(e.clientX, e.clientY);
    }
    
    handleClick(e) {
        // Hide context menu and user menu on click
        this.hideContextMenu();
        this.hideUserMenu();
        
        // Reset auto-lock timer
        this.resetAutoLockTimer();
    }
    
    showContextMenu(x, y) {
        const contextMenu = document.getElementById('context-menu');
        if (!contextMenu) return;
        
        contextMenu.style.display = 'block';
        contextMenu.style.left = `${x}px`;
        contextMenu.style.top = `${y}px`;
        
        // Ensure menu stays within viewport
        const rect = contextMenu.getBoundingClientRect();
        if (rect.right > window.innerWidth) {
            contextMenu.style.left = `${x - rect.width}px`;
        }
        if (rect.bottom > window.innerHeight) {
            contextMenu.style.top = `${y - rect.height}px`;
        }
    }
    
    hideContextMenu() {
        const contextMenu = document.getElementById('context-menu');
        if (contextMenu) contextMenu.style.display = 'none';
    }
    
    showUserMenu() {
        const userMenu = document.getElementById('user-menu');
        if (!userMenu) return;
        
        userMenu.style.display = userMenu.style.display === 'block' ? 'none' : 'block';
    }
    
    hideUserMenu() {
        const userMenu = document.getElementById('user-menu');
        if (userMenu) userMenu.style.display = 'none';
    }
    
    showActivitiesOverview() {
        const overview = document.getElementById('activities-overview');
        if (!overview) return;
        
        overview.style.display = 'flex';
        
        // Focus search input
        const searchInput = document.getElementById('app-search');
        if (searchInput) {
            setTimeout(() => searchInput.focus(), 100);
        }
    }
    
    hideActivitiesOverview() {
        const overview = document.getElementById('activities-overview');
        if (overview) overview.style.display = 'none';
    }
    
    hideOverlays() {
        this.hideActivitiesOverview();
        this.hideContextMenu();
        this.hideUserMenu();
    }
    
    switchWorkspace(number) {
        if (number >= 1 && number <= 3) {
            this.currentWorkspace = number;
            
            // Update workspace indicators
            document.querySelectorAll('.workspace').forEach((ws, index) => {
                ws.classList.toggle('active', index + 1 === number);
            });
            
            // Hide/show windows based on workspace
            this.updateWorkspaceWindows();
            
            this.showNotification('Workspace Changed', `Switched to Workspace ${number}`);
        }
    }
    
    updateWorkspaceWindows() {
        // Implementation for workspace-specific window management
        this.openWindows.forEach((windowData, windowId) => {
            const windowElement = document.getElementById(windowId);
            if (windowElement) {
                const isCurrentWorkspace = windowData.workspace === this.currentWorkspace;
                windowElement.style.display = isCurrentWorkspace ? 'flex' : 'none';
            }
        });
    }
    
    showNotification(title, content, type = 'info', duration = 5000) {
        if (!this.preferences.notifications) return;
        
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-title">${title}</div>
            <div class="notification-content">${content}</div>
        `;
        
        const container = document.getElementById('notifications');
        container.appendChild(notification);
        
        // Auto-remove notification
        setTimeout(() => {
            if (notification.parentNode) {
                notification.style.animation = 'slideOutRight 0.3s ease';
                setTimeout(() => notification.remove(), 300);
            }
        }, duration);
        
        // Limit number of notifications
        const notifications = container.children;
        if (notifications.length > 5) {
            notifications[0].remove();
        }
    }
    
    lockScreen() {
        this.isLocked = true;
        const lockScreen = document.getElementById('lock-screen');
        const desktop = document.getElementById('desktop');
        
        if (lockScreen) lockScreen.style.display = 'flex';
        if (desktop) desktop.style.display = 'none';
        
        // Clear password input
        const passwordInput = document.getElementById('password-input');
        if (passwordInput) {
            passwordInput.value = '';
            setTimeout(() => passwordInput.focus(), 100);
        }
    }
    
    unlock() {
        const passwordInput = document.getElementById('password-input');
        if (!passwordInput) return;
        
        const password = passwordInput.value;
        
        // Simple password check (in real OS this would be properly secured)
        if (password === 'corridor' || password === '') {
            this.isLocked = false;
            const lockScreen = document.getElementById('lock-screen');
            const desktop = document.getElementById('desktop');
            
            if (lockScreen) lockScreen.style.display = 'none';
            if (desktop) desktop.style.display = 'block';
            this.resetAutoLockTimer();
        } else {
            // Shake animation for wrong password
            passwordInput.style.animation = 'shake 0.5s ease';
            setTimeout(() => {
                passwordInput.style.animation = '';
                passwordInput.value = '';
                passwordInput.focus();
            }, 500);
        }
    }
    
    resetAutoLockTimer() {
        if (this.autoLockTimer) {
            clearTimeout(this.autoLockTimer);
        }
        
        if (this.preferences.autoLock > 0 && !this.isLocked) {
            this.autoLockTimer = setTimeout(() => {
                this.lockScreen();
            }, this.preferences.autoLock);
        }
    }
    
    logout() {
        if (confirm('Are you sure you want to log out?')) {
            // Close all windows
            this.openWindows.clear();
            const windowsContainer = document.getElementById('windows-container');
            if (windowsContainer) windowsContainer.innerHTML = '';
            
            // Show boot splash as logout screen
            const desktop = document.getElementById('desktop');
            const bootSplash = document.getElementById('boot-splash');
            const bootStatus = document.getElementById('boot-status');
            const bootProgressFill = document.getElementById('boot-progress-fill');
            
            if (desktop) desktop.style.display = 'none';
            if (bootSplash) bootSplash.style.display = 'flex';
            if (bootStatus) bootStatus.textContent = 'Logging out...';
            if (bootProgressFill) bootProgressFill.style.width = '100%';
            
            // Simulate logout process
            setTimeout(() => {
                location.reload();
            }, 2000);
        }
    }
    
    restart() {
        if (confirm('Are you sure you want to restart?')) {
            this.showNotification('System Restart', 'Restarting Corridor OS...');
            setTimeout(() => location.reload(), 2000);
        }
    }
    
    shutdown() {
        if (confirm('Are you sure you want to shut down?')) {
            // Fade to black
            document.body.style.transition = 'opacity 2s ease';
            document.body.style.opacity = '0';
            
            setTimeout(() => {
                document.body.innerHTML = `
                    <div style="
                        display: flex;
                        align-items: center;
                        justify-content: center;
                        height: 100vh;
                        background: #000;
                        color: #fff;
                        font-family: Ubuntu, sans-serif;
                        flex-direction: column;
                        gap: 20px;
                    ">
                        <div style="font-size: 48px;">⏻</div>
                        <div>Corridor OS has shut down</div>
                        <div style="font-size: 14px; opacity: 0.7;">You can safely close this window</div>
                    </div>
                `;
            }, 2000);
        }
    }
    
    refreshDesktop() {
        // Refresh desktop icons and background
        this.showNotification('Desktop Refreshed', 'Desktop environment has been refreshed.');
    }
    
    toggleFullscreen() {
        if (!document.fullscreenElement) {
            document.documentElement.requestFullscreen();
        } else {
            document.exitFullscreen();
        }
    }
    
    showHelp() {
        this.openApp('help');
    }
    
    // Window management methods will be implemented in corridor-window-manager.js
    handleWindowInteraction(e) {
        // Delegate to window manager
        if (window.corridorWindowManager) {
            window.corridorWindowManager.handleMouseDown(e);
        }
    }
    
    handleMouseMove(e) {
        if (window.corridorWindowManager) {
            window.corridorWindowManager.handleMouseMove(e);
        }
    }
    
    handleMouseUp(e) {
        if (window.corridorWindowManager) {
            window.corridorWindowManager.handleMouseUp(e);
        }
    }
    
    closeActiveWindow() {
        if (window.corridorWindowManager) {
            window.corridorWindowManager.closeActiveWindow();
        }
    }
    
    // App launching will be implemented in corridor-apps.js
    openApp(appName) {
        if (window.corridorApps) {
            window.corridorApps.openApp(appName);
        }
    }
}

// Global functions for HTML event handlers
function showActivitiesOverview() {
    window.corridorOS.showActivitiesOverview();
}

function showUserMenu() {
    window.corridorOS.showUserMenu();
}

function switchWorkspace(number) {
    window.corridorOS.switchWorkspace(number);
}

function lockScreen() {
    window.corridorOS.lockScreen();
}

function unlock() {
    window.corridorOS.unlock();
}

function logout() {
    window.corridorOS.logout();
}

function restart() {
    window.corridorOS.restart();
}

function shutdown() {
    window.corridorOS.shutdown();
}

function openApp(appName) {
    window.corridorOS.openApp(appName);
}

function createFile() {
    window.corridorOS.showNotification('Create File', 'File creation feature coming soon!');
}

function createFolder() {
    window.corridorOS.showNotification('Create Folder', 'Folder creation feature coming soon!');
}

function openTerminalHere() {
    window.corridorOS.openApp('terminal');
}

function openSettings() {
    window.corridorOS.openApp('settings');
}

// Initialize Corridor OS when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.corridorOS = new CorridorOS();
});

// Add CSS for shake animation
const shakeCSS = `
@keyframes shake {
    0%, 100% { transform: translateX(0); }
    10%, 30%, 50%, 70%, 90% { transform: translateX(-10px); }
    20%, 40%, 60%, 80% { transform: translateX(10px); }
}

@keyframes slideOutRight {
    from {
        transform: translateX(0);
        opacity: 1;
    }
    to {
        transform: translateX(100%);
        opacity: 0;
    }
}
`;

const style = document.createElement('style');
style.textContent = shakeCSS;
document.head.appendChild(style);
