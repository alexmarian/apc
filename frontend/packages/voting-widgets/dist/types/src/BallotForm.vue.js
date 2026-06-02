import { watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { NButton, NCard, NCheckbox, NCheckboxGroup, NRadio, NRadioGroup, NSpace, NText } from 'naive-ui';
const { t, locale } = useI18n({
    useScope: 'local',
    messages: {
        en: {
            yes: 'Yes',
            no: 'No',
            abstain: 'Abstain',
            rankingHint: 'Rank options from most preferred (top) to least preferred (bottom).',
        },
        ro: {
            yes: 'Da',
            no: 'Nu',
            abstain: 'Abținere',
            rankingHint: 'Ordonați opțiunile de la cea mai preferată (sus) la cea mai puțin preferată (jos).',
        },
        ru: {
            yes: 'Да',
            no: 'Нет',
            abstain: 'Воздержаться',
            rankingHint: 'Упорядочьте варианты от наиболее предпочтительного (вверху) до наименее предпочтительного (внизу).',
        },
    }
});
const props = defineProps();
const emit = defineEmits();
// Initialise empty entries for matters not yet in modelValue
watch(() => props.matters, (matters) => {
    const current = { ...props.modelValue };
    let changed = false;
    for (const m of matters) {
        const key = String(m.id);
        if (current[key] === undefined) {
            current[key] = m.voting_config.type === 'ranking'
                ? (m.voting_config.options ?? []).map(o => o.id)
                : [];
            changed = true;
        }
    }
    if (changed)
        emit('update:modelValue', current);
}, { immediate: true });
function singleValue(matterId) {
    return props.modelValue[String(matterId)]?.[0] ?? '';
}
function setSingle(matterId, val) {
    emit('update:modelValue', { ...props.modelValue, [String(matterId)]: val ? [val] : [] });
}
function multiValue(matterId) {
    return props.modelValue[String(matterId)] ?? [];
}
function setMulti(matterId, vals) {
    emit('update:modelValue', { ...props.modelValue, [String(matterId)]: vals.map(String) });
}
function rankValue(matterId) {
    return props.modelValue[String(matterId)] ?? [];
}
function moveRank(matterId, idx, dir) {
    const arr = [...rankValue(matterId)];
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= arr.length)
        return;
    [arr[idx], arr[newIdx]] = [arr[newIdx], arr[idx]];
    emit('update:modelValue', { ...props.modelValue, [String(matterId)]: arr });
}
function matterTitle(matter) {
    const lang = locale.value?.slice(0, 2);
    if (lang === 'ru' && matter.title_ru)
        return matter.title_ru;
    return matter.title;
}
function matterDescription(matter) {
    const lang = locale.value?.slice(0, 2);
    if (lang === 'ru' && matter.description_ru)
        return matter.description_ru;
    return matter.description;
}
function optionText(matter, optId) {
    return matter.voting_config.options?.find(o => o.id === optId)?.text ?? optId;
}
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
__VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.matters))) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        key: (matter.id),
        ...{ style: {} },
    });
    const __VLS_0 = {}.NCard;
    /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
    // @ts-ignore
    const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
        size: "small",
    }));
    const __VLS_2 = __VLS_1({
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_1));
    __VLS_3.slots.default;
    {
        const { header: __VLS_thisSlot } = __VLS_3.slots;
        (__VLS_ctx.matterTitle(matter));
    }
    if (matter.description || matter.description_ru) {
        const __VLS_4 = {}.NText;
        /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
        // @ts-ignore
        const __VLS_5 = __VLS_asFunctionalComponent(__VLS_4, new __VLS_4({
            depth: (2),
            ...{ style: {} },
        }));
        const __VLS_6 = __VLS_5({
            depth: (2),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_5));
        __VLS_7.slots.default;
        (__VLS_ctx.matterDescription(matter));
        var __VLS_7;
    }
    if (matter.voting_config.type === 'yes_no') {
        const __VLS_8 = {}.NRadioGroup;
        /** @type {[typeof __VLS_components.NRadioGroup, typeof __VLS_components.NRadioGroup, ]} */ ;
        // @ts-ignore
        const __VLS_9 = __VLS_asFunctionalComponent(__VLS_8, new __VLS_8({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.singleValue(matter.id)),
        }));
        const __VLS_10 = __VLS_9({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.singleValue(matter.id)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_9));
        let __VLS_12;
        let __VLS_13;
        let __VLS_14;
        const __VLS_15 = {
            'onUpdate:value': (...[$event]) => {
                if (!(matter.voting_config.type === 'yes_no'))
                    return;
                __VLS_ctx.setSingle(matter.id, $event);
            }
        };
        __VLS_11.slots.default;
        const __VLS_16 = {}.NSpace;
        /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
        // @ts-ignore
        const __VLS_17 = __VLS_asFunctionalComponent(__VLS_16, new __VLS_16({}));
        const __VLS_18 = __VLS_17({}, ...__VLS_functionalComponentArgsRest(__VLS_17));
        __VLS_19.slots.default;
        const __VLS_20 = {}.NRadio;
        /** @type {[typeof __VLS_components.NRadio, typeof __VLS_components.NRadio, ]} */ ;
        // @ts-ignore
        const __VLS_21 = __VLS_asFunctionalComponent(__VLS_20, new __VLS_20({
            value: "yes",
        }));
        const __VLS_22 = __VLS_21({
            value: "yes",
        }, ...__VLS_functionalComponentArgsRest(__VLS_21));
        __VLS_23.slots.default;
        (__VLS_ctx.t('yes'));
        var __VLS_23;
        const __VLS_24 = {}.NRadio;
        /** @type {[typeof __VLS_components.NRadio, typeof __VLS_components.NRadio, ]} */ ;
        // @ts-ignore
        const __VLS_25 = __VLS_asFunctionalComponent(__VLS_24, new __VLS_24({
            value: "no",
        }));
        const __VLS_26 = __VLS_25({
            value: "no",
        }, ...__VLS_functionalComponentArgsRest(__VLS_25));
        __VLS_27.slots.default;
        (__VLS_ctx.t('no'));
        var __VLS_27;
        if (matter.voting_config.allow_abstention) {
            const __VLS_28 = {}.NRadio;
            /** @type {[typeof __VLS_components.NRadio, typeof __VLS_components.NRadio, ]} */ ;
            // @ts-ignore
            const __VLS_29 = __VLS_asFunctionalComponent(__VLS_28, new __VLS_28({
                value: "abstain",
            }));
            const __VLS_30 = __VLS_29({
                value: "abstain",
            }, ...__VLS_functionalComponentArgsRest(__VLS_29));
            __VLS_31.slots.default;
            (__VLS_ctx.t('abstain'));
            var __VLS_31;
        }
        var __VLS_19;
        var __VLS_11;
    }
    else if (matter.voting_config.type === 'single_choice') {
        const __VLS_32 = {}.NRadioGroup;
        /** @type {[typeof __VLS_components.NRadioGroup, typeof __VLS_components.NRadioGroup, ]} */ ;
        // @ts-ignore
        const __VLS_33 = __VLS_asFunctionalComponent(__VLS_32, new __VLS_32({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.singleValue(matter.id)),
        }));
        const __VLS_34 = __VLS_33({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.singleValue(matter.id)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_33));
        let __VLS_36;
        let __VLS_37;
        let __VLS_38;
        const __VLS_39 = {
            'onUpdate:value': (...[$event]) => {
                if (!!(matter.voting_config.type === 'yes_no'))
                    return;
                if (!(matter.voting_config.type === 'single_choice'))
                    return;
                __VLS_ctx.setSingle(matter.id, $event);
            }
        };
        __VLS_35.slots.default;
        const __VLS_40 = {}.NSpace;
        /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
        // @ts-ignore
        const __VLS_41 = __VLS_asFunctionalComponent(__VLS_40, new __VLS_40({
            vertical: true,
        }));
        const __VLS_42 = __VLS_41({
            vertical: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_41));
        __VLS_43.slots.default;
        for (const [option] of __VLS_getVForSourceType((matter.voting_config.options ?? []))) {
            const __VLS_44 = {}.NRadio;
            /** @type {[typeof __VLS_components.NRadio, typeof __VLS_components.NRadio, ]} */ ;
            // @ts-ignore
            const __VLS_45 = __VLS_asFunctionalComponent(__VLS_44, new __VLS_44({
                key: (option.id),
                value: (option.id),
            }));
            const __VLS_46 = __VLS_45({
                key: (option.id),
                value: (option.id),
            }, ...__VLS_functionalComponentArgsRest(__VLS_45));
            __VLS_47.slots.default;
            (option.text);
            var __VLS_47;
        }
        if (matter.voting_config.allow_abstention) {
            const __VLS_48 = {}.NRadio;
            /** @type {[typeof __VLS_components.NRadio, typeof __VLS_components.NRadio, ]} */ ;
            // @ts-ignore
            const __VLS_49 = __VLS_asFunctionalComponent(__VLS_48, new __VLS_48({
                value: "abstain",
            }));
            const __VLS_50 = __VLS_49({
                value: "abstain",
            }, ...__VLS_functionalComponentArgsRest(__VLS_49));
            __VLS_51.slots.default;
            (__VLS_ctx.t('abstain'));
            var __VLS_51;
        }
        var __VLS_43;
        var __VLS_35;
    }
    else if (matter.voting_config.type === 'multiple_choice') {
        const __VLS_52 = {}.NCheckboxGroup;
        /** @type {[typeof __VLS_components.NCheckboxGroup, typeof __VLS_components.NCheckboxGroup, ]} */ ;
        // @ts-ignore
        const __VLS_53 = __VLS_asFunctionalComponent(__VLS_52, new __VLS_52({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.multiValue(matter.id)),
        }));
        const __VLS_54 = __VLS_53({
            ...{ 'onUpdate:value': {} },
            value: (__VLS_ctx.multiValue(matter.id)),
        }, ...__VLS_functionalComponentArgsRest(__VLS_53));
        let __VLS_56;
        let __VLS_57;
        let __VLS_58;
        const __VLS_59 = {
            'onUpdate:value': (...[$event]) => {
                if (!!(matter.voting_config.type === 'yes_no'))
                    return;
                if (!!(matter.voting_config.type === 'single_choice'))
                    return;
                if (!(matter.voting_config.type === 'multiple_choice'))
                    return;
                __VLS_ctx.setMulti(matter.id, $event);
            }
        };
        __VLS_55.slots.default;
        const __VLS_60 = {}.NSpace;
        /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
        // @ts-ignore
        const __VLS_61 = __VLS_asFunctionalComponent(__VLS_60, new __VLS_60({
            vertical: true,
        }));
        const __VLS_62 = __VLS_61({
            vertical: true,
        }, ...__VLS_functionalComponentArgsRest(__VLS_61));
        __VLS_63.slots.default;
        for (const [option] of __VLS_getVForSourceType((matter.voting_config.options ?? []))) {
            const __VLS_64 = {}.NCheckbox;
            /** @type {[typeof __VLS_components.NCheckbox, typeof __VLS_components.NCheckbox, ]} */ ;
            // @ts-ignore
            const __VLS_65 = __VLS_asFunctionalComponent(__VLS_64, new __VLS_64({
                key: (option.id),
                value: (option.id),
            }));
            const __VLS_66 = __VLS_65({
                key: (option.id),
                value: (option.id),
            }, ...__VLS_functionalComponentArgsRest(__VLS_65));
            __VLS_67.slots.default;
            (option.text);
            var __VLS_67;
        }
        var __VLS_63;
        var __VLS_55;
    }
    else if (matter.voting_config.type === 'ranking') {
        __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({});
        const __VLS_68 = {}.NText;
        /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
        // @ts-ignore
        const __VLS_69 = __VLS_asFunctionalComponent(__VLS_68, new __VLS_68({
            depth: (3),
            ...{ style: {} },
        }));
        const __VLS_70 = __VLS_69({
            depth: (3),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_69));
        __VLS_71.slots.default;
        (__VLS_ctx.t('rankingHint'));
        var __VLS_71;
        for (const [optId, idx] of __VLS_getVForSourceType((__VLS_ctx.rankValue(matter.id)))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                key: (optId),
                ...{ style: {} },
            });
            const __VLS_72 = {}.NText;
            /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
            // @ts-ignore
            const __VLS_73 = __VLS_asFunctionalComponent(__VLS_72, new __VLS_72({
                depth: (3),
                ...{ style: {} },
            }));
            const __VLS_74 = __VLS_73({
                depth: (3),
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_73));
            __VLS_75.slots.default;
            (idx + 1);
            var __VLS_75;
            const __VLS_76 = {}.NText;
            /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
            // @ts-ignore
            const __VLS_77 = __VLS_asFunctionalComponent(__VLS_76, new __VLS_76({
                ...{ style: {} },
            }));
            const __VLS_78 = __VLS_77({
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_77));
            __VLS_79.slots.default;
            (__VLS_ctx.optionText(matter, optId));
            var __VLS_79;
            const __VLS_80 = {}.NButton;
            /** @type {[typeof __VLS_components.NButton, typeof __VLS_components.NButton, ]} */ ;
            // @ts-ignore
            const __VLS_81 = __VLS_asFunctionalComponent(__VLS_80, new __VLS_80({
                ...{ 'onClick': {} },
                size: "tiny",
                disabled: (idx === 0),
            }));
            const __VLS_82 = __VLS_81({
                ...{ 'onClick': {} },
                size: "tiny",
                disabled: (idx === 0),
            }, ...__VLS_functionalComponentArgsRest(__VLS_81));
            let __VLS_84;
            let __VLS_85;
            let __VLS_86;
            const __VLS_87 = {
                onClick: (...[$event]) => {
                    if (!!(matter.voting_config.type === 'yes_no'))
                        return;
                    if (!!(matter.voting_config.type === 'single_choice'))
                        return;
                    if (!!(matter.voting_config.type === 'multiple_choice'))
                        return;
                    if (!(matter.voting_config.type === 'ranking'))
                        return;
                    __VLS_ctx.moveRank(matter.id, idx, -1);
                }
            };
            __VLS_83.slots.default;
            var __VLS_83;
            const __VLS_88 = {}.NButton;
            /** @type {[typeof __VLS_components.NButton, typeof __VLS_components.NButton, ]} */ ;
            // @ts-ignore
            const __VLS_89 = __VLS_asFunctionalComponent(__VLS_88, new __VLS_88({
                ...{ 'onClick': {} },
                size: "tiny",
                disabled: (idx === __VLS_ctx.rankValue(matter.id).length - 1),
            }));
            const __VLS_90 = __VLS_89({
                ...{ 'onClick': {} },
                size: "tiny",
                disabled: (idx === __VLS_ctx.rankValue(matter.id).length - 1),
            }, ...__VLS_functionalComponentArgsRest(__VLS_89));
            let __VLS_92;
            let __VLS_93;
            let __VLS_94;
            const __VLS_95 = {
                onClick: (...[$event]) => {
                    if (!!(matter.voting_config.type === 'yes_no'))
                        return;
                    if (!!(matter.voting_config.type === 'single_choice'))
                        return;
                    if (!!(matter.voting_config.type === 'multiple_choice'))
                        return;
                    if (!(matter.voting_config.type === 'ranking'))
                        return;
                    __VLS_ctx.moveRank(matter.id, idx, 1);
                }
            };
            __VLS_91.slots.default;
            var __VLS_91;
        }
    }
    var __VLS_3;
}
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            NButton: NButton,
            NCard: NCard,
            NCheckbox: NCheckbox,
            NCheckboxGroup: NCheckboxGroup,
            NRadio: NRadio,
            NRadioGroup: NRadioGroup,
            NSpace: NSpace,
            NText: NText,
            t: t,
            singleValue: singleValue,
            setSingle: setSingle,
            multiValue: multiValue,
            setMulti: setMulti,
            rankValue: rankValue,
            moveRank: moveRank,
            matterTitle: matterTitle,
            matterDescription: matterDescription,
            optionText: optionText,
        };
    },
    __typeEmits: {},
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeEmits: {},
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
