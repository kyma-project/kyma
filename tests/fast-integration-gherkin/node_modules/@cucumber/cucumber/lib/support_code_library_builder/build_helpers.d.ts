import { ParameterType } from '@cucumber/cucumber-expressions';
import { ILineAndUri } from '../types';
import { IParameterTypeDefinition } from './types';
export declare function getDefinitionLineAndUri(cwd: string): ILineAndUri;
export declare function buildParameterType({ name, regexp, transformer, useForSnippets, preferForRegexpMatch, }: IParameterTypeDefinition<any>): ParameterType<any>;
