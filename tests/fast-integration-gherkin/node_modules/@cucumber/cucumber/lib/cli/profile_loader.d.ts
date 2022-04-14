export default class ProfileLoader {
    private readonly directory;
    constructor(directory: string);
    getDefinitions(): Promise<Record<string, string>>;
    getArgv(profiles: string[]): Promise<string[]>;
}
