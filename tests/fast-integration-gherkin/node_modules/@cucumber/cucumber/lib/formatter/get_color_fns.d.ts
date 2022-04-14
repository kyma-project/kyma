import { TestStepResultStatus } from '@cucumber/messages';
export declare type IColorFn = (text: string) => string;
export interface IColorFns {
    forStatus: (status: TestStepResultStatus) => IColorFn;
    location: IColorFn;
    tag: IColorFn;
    diffAdded: IColorFn;
    diffRemoved: IColorFn;
    errorMessage: IColorFn;
    errorStack: IColorFn;
}
export default function getColorFns(enabled: boolean): IColorFns;
