import { IParsedArgvFormatOptions } from './argv_parser';
import { IPickleFilterOptions } from '../pickle_filter';
import { IRuntimeOptions } from '../runtime';
export interface IConfigurationFormat {
    outputTo: string;
    type: string;
}
export interface IConfiguration {
    featureDefaultLanguage: string;
    featurePaths: string[];
    formats: IConfigurationFormat[];
    formatOptions: IParsedArgvFormatOptions;
    publishing: boolean;
    listI18nKeywordsFor: string;
    listI18nLanguages: boolean;
    order: string;
    parallel: number;
    pickleFilterOptions: IPickleFilterOptions;
    predictableIds: boolean;
    profiles: string[];
    runtimeOptions: IRuntimeOptions;
    shouldExitImmediately: boolean;
    supportCodePaths: string[];
    supportCodeRequiredModules: string[];
    suppressPublishAdvertisement: boolean;
}
export interface INewConfigurationBuilderOptions {
    argv: string[];
    cwd: string;
}
export default class ConfigurationBuilder {
    static build(options: INewConfigurationBuilderOptions): Promise<IConfiguration>;
    private readonly cwd;
    private readonly args;
    private readonly options;
    constructor({ argv, cwd }: INewConfigurationBuilderOptions);
    build(): Promise<IConfiguration>;
    expandPaths(unexpandedPaths: string[], defaultExtension: string): Promise<string[]>;
    expandFeaturePaths(featurePaths: string[]): Promise<string[]>;
    getFeatureDirectoryPaths(featurePaths: string[]): string[];
    isPublishing(): boolean;
    isPublishAdvertisementSuppressed(): boolean;
    getFormats(): IConfigurationFormat[];
    isTruthyString(s: string | undefined): boolean;
    getUnexpandedFeaturePaths(): Promise<string[]>;
}
