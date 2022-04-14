import { ClassTransformer } from './ClassTransformer';
export { ClassTransformer } from './ClassTransformer';
export * from './decorators';
export * from './interfaces';
export * from './enums';
var classTransformer = new ClassTransformer();
export function classToPlain(object, options) {
    return classTransformer.classToPlain(object, options);
}
export function classToPlainFromExist(object, plainObject, options) {
    return classTransformer.classToPlainFromExist(object, plainObject, options);
}
export function plainToClass(cls, plain, options) {
    return classTransformer.plainToClass(cls, plain, options);
}
export function plainToClassFromExist(clsObject, plain, options) {
    return classTransformer.plainToClassFromExist(clsObject, plain, options);
}
export function classToClass(object, options) {
    return classTransformer.classToClass(object, options);
}
export function classToClassFromExist(object, fromObject, options) {
    return classTransformer.classToClassFromExist(object, fromObject, options);
}
export function serialize(object, options) {
    return classTransformer.serialize(object, options);
}
/**
 * Deserializes given JSON string to a object of the given class.
 */
export function deserialize(cls, json, options) {
    return classTransformer.deserialize(cls, json, options);
}
/**
 * Deserializes given JSON string to an array of objects of the given class.
 */
export function deserializeArray(cls, json, options) {
    return classTransformer.deserializeArray(cls, json, options);
}
//# sourceMappingURL=index.js.map