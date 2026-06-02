import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { NAlert, NCard, NDescriptions, NDescriptionsItem, NDivider, NProgress, NSpace, NSpin, NTag, NText } from 'naive-ui';
const { t } = useI18n({
    useScope: 'local',
    messages: {
        en: {
            gathering: 'Gathering', status: 'Status',
            participationSummary: 'Participation Summary',
            participated: 'Participated', voted: 'Voted', units: 'units',
            participationRate: 'Participation rate',
            passed: 'PASSED', failed: 'FAILED',
            yourVoteCounted: 'Your vote has been counted for this matter.',
            didNotVote: 'You did not vote on this matter.',
            yourVote: 'Your vote', vote: 'vote', votes: 'votes',
            yes: 'Yes', no: 'No', abstain: 'Abstain',
            weight: 'weight',
            quorum: 'Quorum', quorumMet: 'Met', quorumNotMet: 'Not met',
            of: 'of', required: 'required',
            notAvailable: 'Results are not yet available. Current status:',
            willBePublished: 'Results will be published after the gathering is tallied.',
        },
        ro: {
            gathering: 'Adunare', status: 'Status',
            participationSummary: 'Rezumat participare',
            participated: 'Participat', voted: 'Votat', units: 'unități',
            participationRate: 'Rata de participare',
            passed: 'ADOPTAT', failed: 'RESPINS',
            yourVoteCounted: 'Votul dvs. a fost înregistrat pentru acest punct.',
            didNotVote: 'Nu ați votat pentru acest punct.',
            yourVote: 'Votul dvs.', vote: 'vot', votes: 'voturi',
            yes: 'Da', no: 'Nu', abstain: 'Abținere',
            weight: 'pondere',
            quorum: 'Cvorum', quorumMet: 'Întrunit', quorumNotMet: 'Neîntrunit',
            of: 'din', required: 'necesar',
            notAvailable: 'Rezultatele nu sunt disponibile încă. Stare curentă:',
            willBePublished: 'Rezultatele vor fi publicate după numărarea voturilor.',
        },
        ru: {
            gathering: 'Собрание', status: 'Статус',
            participationSummary: 'Сводка участия',
            participated: 'Участвовало', voted: 'Проголосовало', units: 'ед.',
            participationRate: 'Явка',
            passed: 'ПРИНЯТО', failed: 'ОТКЛОНЕНО',
            yourVoteCounted: 'Ваш голос учтён по данному вопросу.',
            didNotVote: 'Вы не голосовали по данному вопросу.',
            yourVote: 'Ваш голос', vote: 'голос', votes: 'голосов',
            yes: 'Да', no: 'Нет', abstain: 'Воздержаться',
            weight: 'вес',
            quorum: 'Кворум', quorumMet: 'Достигнут', quorumNotMet: 'Не достигнут',
            of: 'из', required: 'требуется',
            notAvailable: 'Результаты пока недоступны. Текущий статус:',
            willBePublished: 'Результаты будут опубликованы после подсчёта голосов.',
        },
    }
});
const props = defineProps();
const loading = ref(false);
const fetchError = ref(null);
const context = ref(null);
const results = ref(null);
const statusTagType = computed(() => {
    switch (context.value?.gathering.status) {
        case 'active': return 'success';
        case 'scheduled': return 'info';
        case 'tallied': return 'info';
        case 'closed': return 'error';
        default: return 'default';
    }
});
function memberVotedOnMatter(matterId) {
    if (!context.value?.ballot)
        return false;
    const vote = context.value.ballot.ballot_content[String(matterId)];
    return !!vote && vote.values.length > 0;
}
function isMyChoice(matterId, choice) {
    if (!context.value?.ballot)
        return false;
    const vote = context.value.ballot.ballot_content[String(matterId)];
    return !!vote && vote.values.includes(choice);
}
function formatChoice(choice, config) {
    if (choice === 'abstain')
        return t('abstain');
    if (config.type === 'yes_no')
        return choice === 'yes' ? t('yes') : t('no');
    const opt = config.options?.find(o => o.id === choice);
    return opt ? opt.text : choice;
}
function sortedVotes(matter) {
    return [...matter.votes].sort((a, b) => b.vote_count - a.vote_count);
}
function progressStatus(choice, matter) {
    if (choice === 'abstain')
        return 'warning';
    if (matter.voting_config.type === 'yes_no') {
        if (choice === 'yes')
            return matter.is_passed ? 'success' : 'default';
        if (choice === 'no')
            return matter.is_passed ? 'default' : 'error';
    }
    return 'default';
}
async function fetchContext() {
    if (props.initialContext) {
        context.value = props.initialContext;
        results.value = props.initialContext.results ?? null;
        return;
    }
    loading.value = true;
    fetchError.value = null;
    try {
        const data = await props.service.getContext();
        context.value = data;
        results.value = data.results ?? null;
    }
    catch (err) {
        fetchError.value = err instanceof Error ? err.message : 'Network error';
    }
    finally {
        loading.value = false;
    }
}
onMounted(fetchContext);
debugger; /* PartiallyEnd: #3632/scriptSetup.vue */
const __VLS_ctx = {};
let __VLS_components;
let __VLS_directives;
// CSS variable injection 
// CSS variable injection end 
const __VLS_0 = {}.NSpin;
/** @type {[typeof __VLS_components.NSpin, typeof __VLS_components.NSpin, ]} */ ;
// @ts-ignore
const __VLS_1 = __VLS_asFunctionalComponent(__VLS_0, new __VLS_0({
    show: (__VLS_ctx.loading),
}));
const __VLS_2 = __VLS_1({
    show: (__VLS_ctx.loading),
}, ...__VLS_functionalComponentArgsRest(__VLS_1));
var __VLS_4 = {};
__VLS_3.slots.default;
if (__VLS_ctx.fetchError) {
    const __VLS_5 = {}.NAlert;
    /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
    // @ts-ignore
    const __VLS_6 = __VLS_asFunctionalComponent(__VLS_5, new __VLS_5({
        type: "error",
        ...{ style: {} },
    }));
    const __VLS_7 = __VLS_6({
        type: "error",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    __VLS_8.slots.default;
    (__VLS_ctx.fetchError);
    var __VLS_8;
}
if (__VLS_ctx.context) {
    const __VLS_9 = {}.NCard;
    /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
        ...{ style: {} },
    }));
    const __VLS_11 = __VLS_10({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
    __VLS_12.slots.default;
    const __VLS_13 = {}.NDescriptions;
    /** @type {[typeof __VLS_components.NDescriptions, typeof __VLS_components.NDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
        column: (2),
        labelPlacement: "top",
        size: "small",
    }));
    const __VLS_15 = __VLS_14({
        column: (2),
        labelPlacement: "top",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_14));
    __VLS_16.slots.default;
    const __VLS_17 = {}.NDescriptionsItem;
    /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({
        label: (__VLS_ctx.t('gathering')),
    }));
    const __VLS_19 = __VLS_18({
        label: (__VLS_ctx.t('gathering')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_18));
    __VLS_20.slots.default;
    (__VLS_ctx.context.gathering.title);
    var __VLS_20;
    const __VLS_21 = {}.NDescriptionsItem;
    /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
        label: (__VLS_ctx.t('status')),
    }));
    const __VLS_23 = __VLS_22({
        label: (__VLS_ctx.t('status')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_22));
    __VLS_24.slots.default;
    const __VLS_25 = {}.NTag;
    /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
    // @ts-ignore
    const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
        type: (__VLS_ctx.statusTagType),
        size: "small",
    }));
    const __VLS_27 = __VLS_26({
        type: (__VLS_ctx.statusTagType),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_26));
    __VLS_28.slots.default;
    (__VLS_ctx.context.gathering.status.toUpperCase());
    var __VLS_28;
    var __VLS_24;
    var __VLS_16;
    var __VLS_12;
    if (!__VLS_ctx.results) {
        const __VLS_29 = {}.NAlert;
        /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
        // @ts-ignore
        const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
            type: "info",
        }));
        const __VLS_31 = __VLS_30({
            type: "info",
        }, ...__VLS_functionalComponentArgsRest(__VLS_30));
        __VLS_32.slots.default;
        (__VLS_ctx.t('notAvailable'));
        __VLS_asFunctionalElement(__VLS_intrinsicElements.strong, __VLS_intrinsicElements.strong)({});
        (__VLS_ctx.context.gathering.status);
        if (__VLS_ctx.context.gathering.status !== 'tallied') {
            (__VLS_ctx.t('willBePublished'));
        }
        var __VLS_32;
    }
    else {
        const __VLS_33 = {}.NCard;
        /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
        // @ts-ignore
        const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
            size: "small",
            ...{ style: {} },
            title: (__VLS_ctx.t('participationSummary')),
        }));
        const __VLS_35 = __VLS_34({
            size: "small",
            ...{ style: {} },
            title: (__VLS_ctx.t('participationSummary')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_34));
        __VLS_36.slots.default;
        const __VLS_37 = {}.NDescriptions;
        /** @type {[typeof __VLS_components.NDescriptions, typeof __VLS_components.NDescriptions, ]} */ ;
        // @ts-ignore
        const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({
            column: (3),
            labelPlacement: "top",
            size: "small",
        }));
        const __VLS_39 = __VLS_38({
            column: (3),
            labelPlacement: "top",
            size: "small",
        }, ...__VLS_functionalComponentArgsRest(__VLS_38));
        __VLS_40.slots.default;
        const __VLS_41 = {}.NDescriptionsItem;
        /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
            label: (__VLS_ctx.t('participated')),
        }));
        const __VLS_43 = __VLS_42({
            label: (__VLS_ctx.t('participated')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_42));
        __VLS_44.slots.default;
        (__VLS_ctx.results.statistics.participating_units);
        (__VLS_ctx.t('units'));
        var __VLS_44;
        const __VLS_45 = {}.NDescriptionsItem;
        /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({
            label: (__VLS_ctx.t('voted')),
        }));
        const __VLS_47 = __VLS_46({
            label: (__VLS_ctx.t('voted')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_46));
        __VLS_48.slots.default;
        (__VLS_ctx.results.statistics.voted_units);
        (__VLS_ctx.t('units'));
        var __VLS_48;
        const __VLS_49 = {}.NDescriptionsItem;
        /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
        // @ts-ignore
        const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({
            label: (__VLS_ctx.t('participationRate')),
        }));
        const __VLS_51 = __VLS_50({
            label: (__VLS_ctx.t('participationRate')),
        }, ...__VLS_functionalComponentArgsRest(__VLS_50));
        __VLS_52.slots.default;
        (__VLS_ctx.results.statistics.participation_rate.toFixed(1));
        var __VLS_52;
        var __VLS_40;
        var __VLS_36;
        for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.results.results))) {
            __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                key: (matter.matter_id),
                ...{ style: {} },
            });
            const __VLS_53 = {}.NCard;
            /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
            // @ts-ignore
            const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
                size: "small",
            }));
            const __VLS_55 = __VLS_54({
                size: "small",
            }, ...__VLS_functionalComponentArgsRest(__VLS_54));
            __VLS_56.slots.default;
            {
                const { header: __VLS_thisSlot } = __VLS_56.slots;
                const __VLS_57 = {}.NSpace;
                /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
                // @ts-ignore
                const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
                    align: "center",
                    justify: "space-between",
                    ...{ style: {} },
                }));
                const __VLS_59 = __VLS_58({
                    align: "center",
                    justify: "space-between",
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_58));
                __VLS_60.slots.default;
                const __VLS_61 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
                    strong: true,
                }));
                const __VLS_63 = __VLS_62({
                    strong: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_62));
                __VLS_64.slots.default;
                (matter.matter_title);
                var __VLS_64;
                const __VLS_65 = {}.NTag;
                /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
                // @ts-ignore
                const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
                    type: (matter.is_passed ? 'success' : 'error'),
                    size: "small",
                }));
                const __VLS_67 = __VLS_66({
                    type: (matter.is_passed ? 'success' : 'error'),
                    size: "small",
                }, ...__VLS_functionalComponentArgsRest(__VLS_66));
                __VLS_68.slots.default;
                (matter.is_passed ? __VLS_ctx.t('passed') : __VLS_ctx.t('failed'));
                var __VLS_68;
                var __VLS_60;
            }
            if (__VLS_ctx.memberVotedOnMatter(matter.matter_id)) {
                const __VLS_69 = {}.NAlert;
                /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
                // @ts-ignore
                const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
                    type: "success",
                    size: "small",
                    ...{ style: {} },
                }));
                const __VLS_71 = __VLS_70({
                    type: "success",
                    size: "small",
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_70));
                __VLS_72.slots.default;
                (__VLS_ctx.t('yourVoteCounted'));
                var __VLS_72;
            }
            else if (__VLS_ctx.context.ballot) {
                const __VLS_73 = {}.NAlert;
                /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
                // @ts-ignore
                const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
                    type: "default",
                    size: "small",
                    ...{ style: {} },
                }));
                const __VLS_75 = __VLS_74({
                    type: "default",
                    size: "small",
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_74));
                __VLS_76.slots.default;
                (__VLS_ctx.t('didNotVote'));
                var __VLS_76;
            }
            for (const [vote] of __VLS_getVForSourceType((__VLS_ctx.sortedVotes(matter)))) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                    key: (vote.choice),
                    ...{ style: {} },
                });
                const __VLS_77 = {}.NSpace;
                /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
                // @ts-ignore
                const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
                    align: "center",
                    justify: "space-between",
                    ...{ style: {} },
                }));
                const __VLS_79 = __VLS_78({
                    align: "center",
                    justify: "space-between",
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_78));
                __VLS_80.slots.default;
                const __VLS_81 = {}.NSpace;
                /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
                // @ts-ignore
                const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
                    align: "center",
                    size: "small",
                }));
                const __VLS_83 = __VLS_82({
                    align: "center",
                    size: "small",
                }, ...__VLS_functionalComponentArgsRest(__VLS_82));
                __VLS_84.slots.default;
                const __VLS_85 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
                    ...{ style: (__VLS_ctx.isMyChoice(matter.matter_id, vote.choice) ? 'font-weight:600;color:#18a058' : '') },
                }));
                const __VLS_87 = __VLS_86({
                    ...{ style: (__VLS_ctx.isMyChoice(matter.matter_id, vote.choice) ? 'font-weight:600;color:#18a058' : '') },
                }, ...__VLS_functionalComponentArgsRest(__VLS_86));
                __VLS_88.slots.default;
                (__VLS_ctx.formatChoice(vote.choice, matter.voting_config));
                var __VLS_88;
                if (__VLS_ctx.isMyChoice(matter.matter_id, vote.choice)) {
                    const __VLS_89 = {}.NTag;
                    /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
                    // @ts-ignore
                    const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
                        type: "success",
                        size: "tiny",
                    }));
                    const __VLS_91 = __VLS_90({
                        type: "success",
                        size: "tiny",
                    }, ...__VLS_functionalComponentArgsRest(__VLS_90));
                    __VLS_92.slots.default;
                    (__VLS_ctx.t('yourVote'));
                    var __VLS_92;
                }
                var __VLS_84;
                const __VLS_93 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
                    depth: (2),
                    ...{ style: {} },
                }));
                const __VLS_95 = __VLS_94({
                    depth: (2),
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_94));
                __VLS_96.slots.default;
                (vote.vote_count);
                (vote.vote_count !== 1 ? __VLS_ctx.t('votes') : __VLS_ctx.t('vote'));
                (vote.percentage.toFixed(1));
                if (__VLS_ctx.results?.statistics?.voting_mode === 'by_weight') {
                    (__VLS_ctx.t('weight'));
                    (vote.weight_percentage.toFixed(1));
                }
                var __VLS_96;
                var __VLS_80;
                const __VLS_97 = {}.NProgress;
                /** @type {[typeof __VLS_components.NProgress, ]} */ ;
                // @ts-ignore
                const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
                    type: "line",
                    percentage: (vote.percentage),
                    status: (__VLS_ctx.progressStatus(vote.choice, matter)),
                    showIndicator: (false),
                    height: (8),
                    borderRadius: (4),
                }));
                const __VLS_99 = __VLS_98({
                    type: "line",
                    percentage: (vote.percentage),
                    status: (__VLS_ctx.progressStatus(vote.choice, matter)),
                    showIndicator: (false),
                    height: (8),
                    borderRadius: (4),
                }, ...__VLS_functionalComponentArgsRest(__VLS_98));
            }
            const __VLS_101 = {}.NDivider;
            /** @type {[typeof __VLS_components.NDivider, ]} */ ;
            // @ts-ignore
            const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
                ...{ style: {} },
            }));
            const __VLS_103 = __VLS_102({
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_102));
            const __VLS_105 = {}.NText;
            /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
            // @ts-ignore
            const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
                depth: (3),
                ...{ style: {} },
            }));
            const __VLS_107 = __VLS_106({
                depth: (3),
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_106));
            __VLS_108.slots.default;
            (__VLS_ctx.t('quorum'));
            (matter.quorum_info?.met ? __VLS_ctx.t('quorumMet') : __VLS_ctx.t('quorumNotMet'));
            if (matter.quorum_info) {
                (matter.quorum_info.achieved_percentage.toFixed(1));
                (__VLS_ctx.t('of'));
                (matter.quorum_info.required_percentage);
                (__VLS_ctx.t('required'));
            }
            var __VLS_108;
            var __VLS_56;
        }
    }
}
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            NAlert: NAlert,
            NCard: NCard,
            NDescriptions: NDescriptions,
            NDescriptionsItem: NDescriptionsItem,
            NDivider: NDivider,
            NProgress: NProgress,
            NSpace: NSpace,
            NSpin: NSpin,
            NTag: NTag,
            NText: NText,
            t: t,
            loading: loading,
            fetchError: fetchError,
            context: context,
            results: results,
            statusTagType: statusTagType,
            memberVotedOnMatter: memberVotedOnMatter,
            isMyChoice: isMyChoice,
            formatChoice: formatChoice,
            sortedVotes: sortedVotes,
            progressStatus: progressStatus,
        };
    },
    __typeProps: {},
});
export default (await import('vue')).defineComponent({
    setup() {
        return {};
    },
    __typeProps: {},
});
; /* PartiallyEnd: #4569/main.vue */
