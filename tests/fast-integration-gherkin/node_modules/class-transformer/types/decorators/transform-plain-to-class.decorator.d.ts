import { ClassTransformOptions, ClassConstructor } from '../interfaces';
/**
 * Return the class instance only with the exposed properties.
 *
 * Can be applied to functions and getters/setters only.
 */
export declare function TransformPlainToClass(classType: ClassConstructor<any>, params?: ClassTransformOptions): MethodDecorator;
