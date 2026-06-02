import type { MatterInfo } from './types';
type __VLS_Props = {
    matters: MatterInfo[];
    modelValue: Record<string, string[]>;
};
declare const _default: import("vue").DefineComponent<__VLS_Props, {}, {}, {}, {}, import("vue").ComponentOptionsMixin, import("vue").ComponentOptionsMixin, {
    "update:modelValue": (value: Record<string, string[]>) => any;
}, string, import("vue").PublicProps, Readonly<__VLS_Props> & Readonly<{
    "onUpdate:modelValue"?: ((value: Record<string, string[]>) => any) | undefined;
}>, {}, {}, {}, {}, string, import("vue").ComponentProvideOptions, false, {}, any>;
export default _default;
