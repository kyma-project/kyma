/// <reference types="node" />
import https = require('https');
import request = require('request');
import { Authenticator } from './auth';
import { User } from './config_types';
export interface CredentialStatus {
    readonly token: string;
    readonly clientCertificateData: string;
    readonly clientKeyData: string;
    readonly expirationTimestamp: string;
}
export interface Credential {
    readonly status: CredentialStatus;
}
export declare class ExecAuth implements Authenticator {
    private readonly tokenCache;
    private execFn;
    isAuthProvider(user: User): boolean;
    applyAuthentication(user: User, opts: request.Options | https.RequestOptions): Promise<void>;
    private getToken;
    private getCredential;
}
