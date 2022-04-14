import { ClassTransformer } from '../ClassTransformer';
/**
 * Return the class instance only with the exposed properties.
 *
 * Can be applied to functions and getters/setters only.
 */
export function TransformClassToClass(params) {
    return function (target, propertyKey, descriptor) {
        const classTransformer = new ClassTransformer();
        const originalMethod = descriptor.value;
        descriptor.value = function (...args) {
            const result = originalMethod.apply(this, args);
            const isPromise = !!result && (typeof result === 'object' || typeof result === 'function') && typeof result.then === 'function';
            return isPromise
                ? result.then((data) => classTransformer.classToClass(data, params))
                : classTransformer.classToClass(result, params);
        };
    };
}
//# sourceMappingURL=transform-class-to-class.decorator.js.map