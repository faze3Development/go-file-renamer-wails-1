import { writable } from 'svelte/store';

// Type definitions
export interface ThemeColors {
    '--primary-bg': string;
    '--secondary-bg': string;
    '--card-bg': string;
    '--sidebar-bg': string;
    '--accent-primary': string;
    '--accent-secondary': string;
    '--text-primary': string;
    '--text-secondary': string;
    '--text-muted': string;
    '--border-color': string;
    '--success-color': string;
    '--warning-color': string;
    '--error-color': string;
}

export interface Theme {
    name: string;
    description: string;
    colors: ThemeColors;
}

export interface BackendVerbosity {
    global: boolean;
    watcher: boolean;
    advancedOperations: boolean;
}

export interface AppSettings {
    theme: string;
    showActivityLog: boolean;
    showStats: boolean;
    compactMode: boolean;
    logRetentionLimit: number;
    backendVerbosity: BackendVerbosity;
}

export type SettingsStore = {
    subscribe: (run: (value: AppSettings) => void) => () => void;
    set: (value: AppSettings) => void;
    update: (updater: (value: AppSettings) => AppSettings) => void;
    updateSetting: (key: keyof AppSettings, value: any) => void;
    updateBackendVerbosity: (key: keyof BackendVerbosity, enabled: boolean) => void;
    reset: () => void;
};

// Constants
export const LOG_RETENTION_DEFAULT = 1000;
export const LOG_RETENTION_SOFT_MAX = 5000;
export const LOG_RETENTION_HARD_MAX = 10000;

export const BACKEND_VERBOSITY_DEFAULT: BackendVerbosity = {
    global: false,
    watcher: true,
    advancedOperations: true,
};

// Theme definitions with improved contrast and readability
export const themes: Record<string, Theme> = {
    default: {
        name: 'Starlight', // Default theme
        description: 'Deep space with vibrant purple and blue nebulas.',
        colors: {
            '--primary-bg': '#0D1117',
            '--secondary-bg': '#161B22',
            '--card-bg': '#0D1117',
            '--sidebar-bg': '#010409',
            '--accent-primary': '#8b5cf6', // Purple
            '--accent-secondary': '#3b82f6', // Blue
            '--text-primary': '#e6edf3',
            '--text-secondary': '#7d8590',
            '--text-muted': '#586069',
            '--border-color': '#30363d',
            '--success-color': '#238636',
            '--warning-color': '#d29922',
            '--error-color': '#f85149',
        }
    },
    cyberpunk: {
        name: 'Cyberpunk',
        description: 'High-tech, low-life. Neon yellow and cyan.',
        colors: {
            '--primary-bg': '#000000',
            '--secondary-bg': '#0c0c0c',
            '--card-bg': '#050505',
            '--sidebar-bg': '#000000',
            '--accent-primary': '#fcee0a', // Yellow
            '--accent-secondary': '#00f0ff', // Cyan
            '--text-primary': '#ffffff',
            '--text-secondary': '#b0b0b0',
            '--text-muted': '#6a6a6a',
            '--border-color': '#333333',
            '--success-color': '#00ff7f',
            '--warning-color': '#ffae00',
            '--error-color': '#ff003c',
        }
    },
    forest: {
        name: 'Forest',
        description: 'Earthy greens and warm, natural tones.',
        colors: {
            '--primary-bg': '#1a201a',
            '--secondary-bg': '#202820',
            '--card-bg': '#1a201a',
            '--sidebar-bg': '#101410',
            '--accent-primary': '#4ade80', // Green
            '--accent-secondary': '#f97316', // Orange
            '--text-primary': '#f0fdf4',
            '--text-secondary': '#a3a3a3',
            '--text-muted': '#6b7280',
            '--border-color': '#374151',
            '--success-color': '#10b981',
            '--warning-color': '#f59e0b',
            '--error-color': '#ef4444',
        }
    },
};

const defaultSettings: AppSettings = {
    theme: 'default',
    showActivityLog: true,
    showStats: true,
    compactMode: false,
    logRetentionLimit: LOG_RETENTION_DEFAULT,
    backendVerbosity: { ...BACKEND_VERBOSITY_DEFAULT },
};

// Settings Store - Restored original API
function sanitizeLogRetention(value: unknown): number {
    const numeric = Number(value);
    if (!Number.isFinite(numeric) || numeric <= 0) {
        return LOG_RETENTION_DEFAULT;
    }
    const floored = Math.floor(numeric);
    return Math.min(Math.max(floored, 100), LOG_RETENTION_HARD_MAX);
}

function sanitizeBackendVerbosity(value: unknown): BackendVerbosity {
    const sanitized = { ...BACKEND_VERBOSITY_DEFAULT };
    if (value && typeof value === 'object') {
        Object.keys(sanitized).forEach((key) => {
            if (Object.prototype.hasOwnProperty.call(value, key)) {
                const verbosityKey = key as keyof BackendVerbosity;
                sanitized[verbosityKey] = Boolean((value as any)[key]);
            }
        });
    }
    return sanitized;
}

function createSettings(): SettingsStore {
    // Load settings from localStorage
    const stored = typeof localStorage !== 'undefined' ? localStorage.getItem('fileRenamerSettings') : null;
    let initial: AppSettings = stored ? { ...defaultSettings, ...JSON.parse(stored) } : { ...defaultSettings };
    initial.logRetentionLimit = sanitizeLogRetention(initial.logRetentionLimit);
    initial.backendVerbosity = sanitizeBackendVerbosity(initial.backendVerbosity);

    const { subscribe, set, update } = writable<AppSettings>(initial);

    return {
        subscribe,
        set,
        update,
        updateSetting: (key: keyof AppSettings, value: any) => {
            update(settings => {
                let nextValue: any = value;

                if (key === 'logRetentionLimit') {
                    nextValue = sanitizeLogRetention(value);
                } else if (key === 'backendVerbosity') {
                    nextValue = sanitizeBackendVerbosity(value);
                }

                const newSettings: AppSettings = { ...settings, [key]: nextValue };
                if (typeof localStorage !== 'undefined') {
                    localStorage.setItem('fileRenamerSettings', JSON.stringify(newSettings));
                }
                if (key === 'theme') {
                    applyTheme(value);
                }
                return newSettings;
            });
        },
        updateBackendVerbosity: (key: keyof BackendVerbosity, enabled: boolean) => {
            update(settings => {
                const currentVerbosity = sanitizeBackendVerbosity(settings.backendVerbosity);
                if (!Object.prototype.hasOwnProperty.call(currentVerbosity, key)) {
                    return settings;
                }

                const nextVerbosity: BackendVerbosity = { ...currentVerbosity, [key]: Boolean(enabled) };
                const newSettings: AppSettings = { ...settings, backendVerbosity: nextVerbosity };

                if (typeof localStorage !== 'undefined') {
                    localStorage.setItem('fileRenamerSettings', JSON.stringify(newSettings));
                }
                return newSettings;
            });
        },
        reset: () => {
            const resetValues: AppSettings = {
                ...defaultSettings,
                backendVerbosity: { ...BACKEND_VERBOSITY_DEFAULT },
            };
            set(resetValues);
            if (typeof localStorage !== 'undefined') {
                localStorage.setItem('fileRenamerSettings', JSON.stringify(resetValues));
            }
            applyTheme(resetValues.theme);
        }
    };
}

export const settings = createSettings();

// Apply theme to document with error handling
export function applyTheme(themeName: string): void {
    try {
        const theme = themes[themeName];
        if (!theme) {
            console.warn(`Theme "${themeName}" not found, falling back to default theme`);
            return applyTheme('default');
        }

        if (typeof document === 'undefined') {
            console.warn('Document not available for theme application');
            return;
        }

        const root = document.documentElement;
        if (!root) {
            console.warn('Document root not available');
            return;
        }

        // Apply theme colors with error handling for each property
        Object.entries(theme.colors).forEach(([property, value]) => {
            try {
                root.style.setProperty(property, value);
            } catch (error) {
                console.warn(`Failed to set CSS property ${property}:`, error);
            }
        });

        console.log(`Theme "${themeName}" applied successfully`);
    } catch (error) {
        console.error('Failed to apply theme:', error);
        // Try to fall back to a safe theme
        if (themeName !== 'default') {
            console.log('Falling back to default theme');
            applyTheme('default');
        }
    }
}