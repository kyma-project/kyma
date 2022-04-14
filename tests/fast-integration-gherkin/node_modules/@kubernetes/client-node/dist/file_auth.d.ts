/// <reference types="node" />
import https = require('https');
import request = require('request');
import { Authenticator } from './auth';
import { User } from './config_types';
export declare class FileAuth implements Authenticator {
    private token;
    private lastRead;
    isAuthProvider(user: User): boolean;
    applyAuthentication(user: User, opts: request.Options | https.RequestOptions): Promise<void>;
    private refreshToken;
    private isTokenExpired;
}
