import { ClassTransformer } from '../ClassTransformer';
/**
 * Transform the object from class to plain object and return only with the exposed properties.
 *
 * Can be applied to functions and getters/setters only.
 */
export function TransformClassToPlain(params) {
    return function (target, propertyKey, descriptor) {
        const classTransformer = new ClassTransformer();
        const originalMethod = descriptor.value;
        descriptor.value = function (...args) {
            const result = originalMethod.apply(this, args);
            const isPromise = !!result && (typeof result === 'object' || typeof result === 'function') && typeof result.then === 'function';
            return isPromise
                ? result.then((data) => classTransformer.classToPlain(data, params))
                : classTransformer.classToPlain(result, params);
        };
    };
}
//# sourceMappingURL=transform-class-to-plain.decorator.js.map