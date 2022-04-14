import { ClassTransformer } from '../ClassTransformer';
/**
 * Return the class instance only with the exposed properties.
 *
 * Can be applied to functions and getters/setters only.
 */
export function TransformPlainToClass(classType, params) {
    return function (target, propertyKey, descriptor) {
        const classTransformer = new ClassTransformer();
        const originalMethod = descriptor.value;
        descriptor.value = function (...args) {
            const result = originalMethod.apply(this, args);
            const isPromise = !!result && (typeof result === 'object' || typeof result === 'function') && typeof result.then === 'function';
            return isPromise
                ? result.then((data) => classTransformer.plainToClass(classType, data, params))
                : classTransformer.plainToClass(classType, result, params);
        };
    };
}
//# sourceMappingURL=transform-plain-to-class.decorator.js.map