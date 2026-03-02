// Type declarations for Wails runtime
declare global {
    interface Window {
        runtime: {
            EventsOnMultiple: (eventName: string, callback: (...data: any) => void, maxCallbacks: number) => () => void;
            EventsOn: (eventName: string, callback: (...data: any) => void) => () => void;
            EventsEmit: (eventName: string, ...data: any) => void;
            EventsOff: (eventName: string, ...additionalEventNames: string[]) => void;
            EventsOffAll: () => void;
            LogPrint: (message: string) => void;
            LogTrace: (message: string) => void;
            LogDebug: (message: string) => void;
            LogError: (message: string) => void;
            LogFatal: (message: string) => void;
            LogInfo: (message: string) => void;
            LogWarning: (message: string) => void;
            WindowReload: () => void;
            WindowReloadApp: () => void;
            WindowSetAlwaysOnTop: (b: boolean) => void;
            WindowSetSystemDefaultTheme: () => void;
            WindowSetLightTheme: () => void;
            WindowSetDarkTheme: () => void;
            WindowCenter: () => void;
            WindowSetTitle: (title: string) => void;
            WindowFullscreen: () => void;
            WindowUnfullscreen: () => void;
            WindowIsFullscreen: () => Promise<boolean>;
            WindowSetSize: (width: number, height: number) => void;
            WindowGetSize: () => Promise<{ w: number; h: number }>;
            WindowSetMaxSize: (width: number, height: number) => void;
            WindowSetMinSize: (width: number, height: number) => void;
            WindowSetPosition: (x: number, y: number) => void;
            WindowGetPosition: () => Promise<{ x: number; y: number }>;
            WindowHide: () => void;
            WindowShow: () => void;
            WindowMaximise: () => void;
            WindowToggleMaximise: () => void;
            WindowUnmaximise: () => void;
            WindowIsMaximised: () => Promise<boolean>;
            WindowMinimise: () => void;
            WindowUnminimise: () => void;
            WindowIsMinimised: () => Promise<boolean>;
            WindowIsNormal: () => Promise<boolean>;
            WindowSetBackgroundColour: (R: number, G: number, B: number, A: number) => void;
            ScreenGetAll: () => Promise<Array<{ isCurrent: boolean; isPrimary: boolean; width: number; height: number }>>;
            BrowserOpenURL: (url: string) => void;
            Environment: () => Promise<{ buildType: string; platform: string; arch: string }>;
            Quit: () => void;
            Hide: () => void;
            Show: () => void;
            ClipboardGetText: () => Promise<string>;
            ClipboardSetText: (text: string) => Promise<boolean>;
            OnFileDrop: (callback: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean) => void;
            OnFileDropOff: () => void;
            CanResolveFilePaths: () => boolean;
            ResolveFilePaths: (files: File[]) => void;
        };
        go: {
            main: {
                DeleteProfile: (arg1: string) => Promise<void>;
                GetActionInfo: () => Promise<Array<{
                    id: string;
                    name: string;
                    description: string;
                    fieldLabel: string;
                }>>;
                GetNamerInfo: () => Promise<Array<{
                    id: string;
                    name: string;
                    description: string;
                }>>;
                GetPatternInfo: () => Promise<Array<{
                    id: string;
                    name: string;
                    description: string;
                    regex: string;
                }>>;
                LoadProfiles: () => Promise<Record<string, {
                    WatchPaths: string[];
                    Recursive: boolean;
                    DryRun: boolean;
                    NamePattern: string;
                    RandomLength: number;
                    Settle: number;
                    SettleTimeout: number;
                    Retries: number;
                    NoInitialScan: boolean;
                    NamerID: string;
                    ActionID: string;
                    TemplateString: string;
                    DateTimeFormat: string;
                    ActionConfig: Record<string, string>;
                }>>;
                SaveProfile: (arg1: string, arg2: {
                    WatchPaths: string[];
                    Recursive: boolean;
                    DryRun: boolean;
                    NamePattern: string;
                    RandomLength: number;
                    Settle: number;
                    SettleTimeout: number;
                    Retries: number;
                    NoInitialScan: boolean;
                    NamerID: string;
                    ActionID: string;
                    TemplateString: string;
                    DateTimeFormat: string;
                    ActionConfig: Record<string, string>;
                }) => Promise<void>;
                SelectActionDirectory: () => Promise<string>;
                SelectDirectory: () => Promise<string>;
                StartWatching: (arg1: {
                    WatchPaths: string[];
                    Recursive: boolean;
                    DryRun: boolean;
                    NamePattern: string;
                    RandomLength: number;
                    Settle: number;
                    SettleTimeout: number;
                    Retries: number;
                    NoInitialScan: boolean;
                    NamerID: string;
                    ActionID: string;
                    TemplateString: string;
                    DateTimeFormat: string;
                    ActionConfig: Record<string, string>;
                }) => Promise<void>;
                StopWatching: () => Promise<void>;
            };
        };
    }
}

export { };
