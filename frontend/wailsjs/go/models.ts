export namespace action {
	
	export class Info {
	    id: string;
	    name: string;
	    description: string;
	    fieldLabel: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.fieldLabel = source["fieldLabel"];
	    }
	}

}

export namespace advanced_file_operations {
	
	export class BulkProcessingFile {
	    filename: string;
	    contentBase64: string;
	    contentType?: string;
	    size: number;
	
	    static createFrom(source: any = {}) {
	        return new BulkProcessingFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.contentBase64 = source["contentBase64"];
	        this.contentType = source["contentType"];
	        this.size = source["size"];
	    }
	}
	export class BulkProcessingOptions {
	    renameFiles: boolean;
	    removeMetadata: boolean;
	    optimizeFiles: boolean;
	    compressFiles: boolean;
	    pattern?: string;
	    namer?: string;
	    renameOptions: bulk_file_processing.RenameOperationOptions;
	    allowedTypes?: string[];
	    maxFileSize?: number;
	
	    static createFrom(source: any = {}) {
	        return new BulkProcessingOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.renameFiles = source["renameFiles"];
	        this.removeMetadata = source["removeMetadata"];
	        this.optimizeFiles = source["optimizeFiles"];
	        this.compressFiles = source["compressFiles"];
	        this.pattern = source["pattern"];
	        this.namer = source["namer"];
	        this.renameOptions = this.convertValues(source["renameOptions"], bulk_file_processing.RenameOperationOptions);
	        this.allowedTypes = source["allowedTypes"];
	        this.maxFileSize = source["maxFileSize"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BulkProcessingRequest {
	    userId?: string;
	    files: BulkProcessingFile[];
	    options: BulkProcessingOptions;
	
	    static createFrom(source: any = {}) {
	        return new BulkProcessingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userId = source["userId"];
	        this.files = this.convertValues(source["files"], BulkProcessingFile);
	        this.options = this.convertValues(source["options"], BulkProcessingOptions);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BulkProcessingResultFile {
	    filename: string;
	    newName?: string;
	    success: boolean;
	    error?: string;
	    action?: string;
	    contentType?: string;
	    contentBase64?: string;
	
	    static createFrom(source: any = {}) {
	        return new BulkProcessingResultFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.newName = source["newName"];
	        this.success = source["success"];
	        this.error = source["error"];
	        this.action = source["action"];
	        this.contentType = source["contentType"];
	        this.contentBase64 = source["contentBase64"];
	    }
	}
	export class BulkProcessingResponse {
	    jobId: string;
	    totalFiles: number;
	    successCount: number;
	    failureCount: number;
	    durationMs: number;
	    results: BulkProcessingResultFile[];
	
	    static createFrom(source: any = {}) {
	        return new BulkProcessingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobId = source["jobId"];
	        this.totalFiles = source["totalFiles"];
	        this.successCount = source["successCount"];
	        this.failureCount = source["failureCount"];
	        this.durationMs = source["durationMs"];
	        this.results = this.convertValues(source["results"], BulkProcessingResultFile);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace bulk_file_processing {
	
	export class FileProcessingResult {
	    filename: string;
	    newName?: string;
	    success: boolean;
	    action: string;
	    error?: string;
	    contentType?: string;
	
	    static createFrom(source: any = {}) {
	        return new FileProcessingResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filename = source["filename"];
	        this.newName = source["newName"];
	        this.success = source["success"];
	        this.action = source["action"];
	        this.error = source["error"];
	        this.contentType = source["contentType"];
	    }
	}
	export class FileUploadMetadata {
	    Filename: string;
	    ContentType: string;
	    Size: number;
	
	    static createFrom(source: any = {}) {
	        return new FileUploadMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Filename = source["Filename"];
	        this.ContentType = source["ContentType"];
	        this.Size = source["Size"];
	    }
	}
	export class SequentialNamingOptions {
	    enabled: boolean;
	    baseName: string;
	    startIndex: number;
	    padLength: number;
	    keepExtension: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SequentialNamingOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.enabled = source["enabled"];
	        this.baseName = source["baseName"];
	        this.startIndex = source["startIndex"];
	        this.padLength = source["padLength"];
	        this.keepExtension = source["keepExtension"];
	    }
	}
	export class RenameOperationOptions {
	    preserveOriginalName: boolean;
	    addTimestamp: boolean;
	    addRandomId: boolean;
	    addCustomDate: boolean;
	    customDate?: string;
	    useRegexReplace: boolean;
	    regexFind?: string;
	    regexReplace?: string;
	    sequentialNaming: SequentialNamingOptions;
	
	    static createFrom(source: any = {}) {
	        return new RenameOperationOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.preserveOriginalName = source["preserveOriginalName"];
	        this.addTimestamp = source["addTimestamp"];
	        this.addRandomId = source["addRandomId"];
	        this.addCustomDate = source["addCustomDate"];
	        this.customDate = source["customDate"];
	        this.useRegexReplace = source["useRegexReplace"];
	        this.regexFind = source["regexFind"];
	        this.regexReplace = source["regexReplace"];
	        this.sequentialNaming = this.convertValues(source["sequentialNaming"], SequentialNamingOptions);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProcessingOptions {
	    renameFiles: boolean;
	    removeMetadata: boolean;
	    compressFiles: boolean;
	    optimizeFiles: boolean;
	    pattern: string;
	    namer: string;
	    renameOptions: RenameOperationOptions;
	    maxFileSize: number;
	    allowedTypes: string[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessingOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.renameFiles = source["renameFiles"];
	        this.removeMetadata = source["removeMetadata"];
	        this.compressFiles = source["compressFiles"];
	        this.optimizeFiles = source["optimizeFiles"];
	        this.pattern = source["pattern"];
	        this.namer = source["namer"];
	        this.renameOptions = this.convertValues(source["renameOptions"], RenameOperationOptions);
	        this.maxFileSize = source["maxFileSize"];
	        this.allowedTypes = source["allowedTypes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProcessingJob {
	    id: string;
	    userId: string;
	    status: string;
	    files: FileUploadMetadata[];
	    options: ProcessingOptions;
	    results: FileProcessingResult[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    startedAt?: any;
	    // Go type: time
	    completedAt?: any;
	    durationMs?: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessingJob(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.userId = source["userId"];
	        this.status = source["status"];
	        this.files = this.convertValues(source["files"], FileUploadMetadata);
	        this.options = this.convertValues(source["options"], ProcessingOptions);
	        this.results = this.convertValues(source["results"], FileProcessingResult);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.completedAt = this.convertValues(source["completedAt"], null);
	        this.durationMs = source["durationMs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

export namespace config {
	
	export class Config {
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
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.WatchPaths = source["WatchPaths"];
	        this.Recursive = source["Recursive"];
	        this.DryRun = source["DryRun"];
	        this.NamePattern = source["NamePattern"];
	        this.RandomLength = source["RandomLength"];
	        this.Settle = source["Settle"];
	        this.SettleTimeout = source["SettleTimeout"];
	        this.Retries = source["Retries"];
	        this.NoInitialScan = source["NoInitialScan"];
	        this.NamerID = source["NamerID"];
	        this.ActionID = source["ActionID"];
	        this.TemplateString = source["TemplateString"];
	        this.DateTimeFormat = source["DateTimeFormat"];
	        this.ActionConfig = source["ActionConfig"];
	    }
	}

}

export namespace namer {
	
	export class Info {
	    id: string;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Info(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}

}

export namespace patterns {
	
	export class PatternInfo {
	    id: string;
	    name: string;
	    description: string;
	    regex: string;
	
	    static createFrom(source: any = {}) {
	        return new PatternInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.regex = source["regex"];
	    }
	}

}

