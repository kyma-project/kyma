import { ClassTransformOptions } from './interfaces';
import { ClassConstructor } from './interfaces';
export { ClassTransformer } from './ClassTransformer';
export * from './decorators';
export * from './interfaces';
export * from './enums';
/**
 * Converts class (constructor) object to plain (literal) object. Also works with arrays.
 */
export declare function classToPlain<T>(object: T, options?: ClassTransformOptions): Record<string, any>;
export declare function classToPlain<T>(object: T[], options?: ClassTransformOptions): Record<string, any>[];
/**
 * Converts class (constructor) object to plain (literal) object.
 * Uses given plain object as source object (it means fills given plain object with data from class object).
 * Also works with arrays.
 */
export declare function classToPlainFromExist<T>(object: T, plainObject: Record<string, any>, options?: ClassTransformOptions): Record<string, any>;
export declare function classToPlainFromExist<T>(object: T, plainObjects: Record<string, any>[], options?: ClassTransformOptions): Record<string, any>[];
/**
 * Converts plain (literal) object to class (constructor) object. Also works with arrays.
 */
export declare function plainToClass<T, V>(cls: ClassConstructor<T>, plain: V[], options?: ClassTransformOptions): T[];
export declare function plainToClass<T, V>(cls: ClassConstructor<T>, plain: V, options?: ClassTransformOptions): T;
/**
 * Converts plain (literal) object to class (constructor) object.
 * Uses given object as source object (it means fills given object with data from plain object).
 *  Also works with arrays.
 */
export declare function plainToClassFromExist<T, V>(clsObject: T[], plain: V[], options?: ClassTransformOptions): T[];
export declare function plainToClassFromExist<T, V>(clsObject: T, plain: V, options?: ClassTransformOptions): T;
/**
 * Converts class (constructor) object to new class (constructor) object. Also works with arrays.
 */
export declare function classToClass<T>(object: T, options?: ClassTransformOptions): T;
export declare function classToClass<T>(object: T[], options?: ClassTransformOptions): T[];
/**
 * Converts class (constructor) object to plain (literal) object.
 * Uses given plain object as source object (it means fills given plain object with data from class object).
 * Also works with arrays.
 */
export declare function classToClassFromExist<T>(object: T, fromObject: T, options?: ClassTransformOptions): T;
export declare function classToClassFromExist<T>(object: T, fromObjects: T[], options?: ClassTransformOptions): T[];
/**
 * Serializes given object to a JSON string.
 */
export declare function serialize<T>(object: T, options?: ClassTransformOptions): string;
export declare function serialize<T>(object: T[], options?: ClassTransformOptions): string;
/**
 * Deserializes given JSON string to a object of the given class.
 */
export declare function deserialize<T>(cls: ClassConstructor<T>, json: string, options?: ClassTransformOptions): T;
/**
 * Deserializes given JSON string to an array of objects of the given class.
 */
export declare function deserializeArray<T>(cls: ClassConstructor<T>, json: string, options?: ClassTransformOptions): T[];
