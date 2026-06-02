import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { NAlert, NButton, NCard, NDescriptions, NDescriptionsItem, NDivider, NSpace, NSpin, NTag, NText } from 'naive-ui';
import BallotForm from './BallotForm.vue';
import { HttpError } from './types';
const { t, locale } = useI18n({
    useScope: 'local',
    messages: {
        en: {
            owner: 'Owner',
            units: 'Units',
            votingWeight: 'Voting weight (%)',
            votingUnits: 'Voting units',
            yourBallot: 'Your Ballot',
            submitBallot: 'Submit Ballot',
            ballotSubmitted: 'Ballot Submitted',
            ballotId: 'Ballot ID',
            verificationHash: 'Verification hash',
            submittedAt: 'Submitted at',
            yourVotes: 'Your votes',
            informational: 'Informational',
            yes: 'Yes', no: 'No', abstain: 'Abstain',
            statusNotStarted: 'Voting has not started yet. Please check back later.',
            statusTallied: 'Voting has closed. Results are being tallied.',
            statusClosed: 'Voting is closed.',
            errAlreadySubmitted: 'A ballot has already been submitted for this gathering.',
            errInvalidBallot: 'Invalid ballot.',
            unitsAlreadyVoted: 'The voting rights for your unit(s) have already been exercised by another co-owner.',
        },
        ro: {
            owner: 'Proprietar',
            units: 'Unități',
            votingWeight: 'Pondere de vot (%)',
            votingUnits: 'Unități de vot',
            yourBallot: 'Buletinul dvs. de vot',
            submitBallot: 'Trimite buletinul',
            ballotSubmitted: 'Buletin trimis',
            ballotId: 'ID buletin',
            verificationHash: 'Hash de verificare',
            submittedAt: 'Trimis la',
            yourVotes: 'Voturile dvs.',
            informational: 'Informativ',
            yes: 'Da', no: 'Nu', abstain: 'Abținere',
            statusNotStarted: 'Votul nu a început încă. Verificați mai târziu.',
            statusTallied: 'Votul s-a încheiat. Rezultatele sunt în curs de numărare.',
            statusClosed: 'Votul este închis.',
            errAlreadySubmitted: 'Un buletin de vot a fost deja trimis pentru această adunare.',
            errInvalidBallot: 'Buletin de vot invalid.',
            unitsAlreadyVoted: 'Dreptul de vot pentru unitatea dvs. a fost deja exercitat de un alt coproprietar.',
        },
        ru: {
            owner: 'Владелец',
            units: 'Единицы',
            votingWeight: 'Вес голоса (%)',
            votingUnits: 'Единицы голосования',
            yourBallot: 'Ваш бюллетень',
            submitBallot: 'Подать бюллетень',
            ballotSubmitted: 'Бюллетень подан',
            ballotId: 'ID бюллетеня',
            verificationHash: 'Хэш верификации',
            submittedAt: 'Подан в',
            yourVotes: 'Ваши голоса',
            informational: 'Информационный',
            yes: 'Да', no: 'Нет', abstain: 'Воздержаться',
            statusNotStarted: 'Голосование ещё не началось. Загляните позже.',
            statusTallied: 'Голосование завершено. Результаты подсчитываются.',
            statusClosed: 'Голосование закрыто.',
            errAlreadySubmitted: 'Бюллетень для этого собрания уже был подан.',
            errInvalidBallot: 'Недействительный бюллетень.',
            unitsAlreadyVoted: 'Право голоса за вашу единицу уже было реализовано другим совладельцем.',
        },
    }
});
const props = defineProps();
const loading = ref(false);
const submitting = ref(false);
const fetchError = ref(null);
const submitError = ref(null);
const context = ref(null);
const receipt = ref(null);
const ballotVotes = ref({});
const ownerPartSum = computed(() => context.value?.units.reduce((sum, u) => sum + u.voting_weight, 0) ?? 0);
const votingShare = computed(() => {
    const g = context.value?.gathering;
    if (!g)
        return 0;
    if (g.voting_mode === 'by_weight') {
        const denom = g.qualified_units_total_part;
        return denom > 0 ? (ownerPartSum.value / denom) * 100 : 0;
    }
    else {
        const denom = g.qualified_units_count;
        return denom > 0 ? ((context.value?.units.length ?? 0) / denom) * 100 : 0;
    }
});
const allUnitsUnavailable = computed(() => (context.value?.units.length ?? 0) > 0 &&
    context.value.units.every(u => !u.is_available));
const votableMatters = computed(() => (context.value?.matters ?? []).filter(m => !m.is_informative));
const informativeMatters = computed(() => (context.value?.matters ?? []).filter(m => m.is_informative));
const canSubmit = computed(() => votableMatters.value.length > 0 &&
    votableMatters.value.every(m => (ballotVotes.value[String(m.id)]?.length ?? 0) > 0));
const statusTagType = computed(() => {
    switch (context.value?.gathering.status) {
        case 'active': return 'success';
        case 'scheduled': return 'info';
        case 'tallied': return 'info';
        case 'closed': return 'error';
        default: return 'default';
    }
});
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
function buildBallotContent() {
    const content = {};
    for (const m of votableMatters.value) {
        const key = String(m.id);
        content[key] = { matter_id: m.id, values: ballotVotes.value[key] ?? [] };
    }
    return content;
}
function formatVotedValues(matter) {
    if (!receipt.value)
        return '—';
    const vote = receipt.value.ballot_content[String(matter.id)];
    if (!vote || vote.values.length === 0)
        return '—';
    if (matter.voting_config.type === 'ranking') {
        return vote.values
            .map((v, i) => {
            const opt = matter.voting_config.options?.find(o => o.id === v);
            return `${i + 1}. ${opt ? opt.text : v}`;
        })
            .join(', ');
    }
    return vote.values
        .map(v => {
        if (v === 'abstain')
            return t('abstain');
        if (matter.voting_config.type === 'yes_no')
            return v === 'yes' ? t('yes') : t('no');
        const opt = matter.voting_config.options?.find(o => o.id === v);
        return opt ? opt.text : v;
    })
        .join(', ');
}
async function fetchContext() {
    if (props.initialContext) {
        context.value = props.initialContext;
        if (props.initialContext.ballot) {
            receipt.value = {
                ballot_id: props.initialContext.ballot.ballot_id,
                ballot_hash: props.initialContext.ballot.ballot_hash,
                submitted_at: props.initialContext.ballot.submitted_at,
                ballot_content: props.initialContext.ballot.ballot_content
            };
        }
        return;
    }
    loading.value = true;
    fetchError.value = null;
    try {
        const data = await props.service.getContext();
        context.value = data;
        if (data.ballot) {
            receipt.value = {
                ballot_id: data.ballot.ballot_id,
                ballot_hash: data.ballot.ballot_hash,
                submitted_at: data.ballot.submitted_at,
                ballot_content: data.ballot.ballot_content
            };
        }
    }
    catch (err) {
        fetchError.value = err instanceof Error ? err.message : 'Network error';
    }
    finally {
        loading.value = false;
    }
}
async function handleSubmit() {
    if (!canSubmit.value)
        return;
    submitting.value = true;
    submitError.value = null;
    const ballotContent = buildBallotContent();
    try {
        const data = await props.service.submitBallot(ballotContent);
        receipt.value = {
            ballot_id: data.ballot_id,
            ballot_hash: data.ballot_hash,
            submitted_at: data.submitted_at,
            ballot_content: data.ballot_content ?? ballotContent
        };
    }
    catch (err) {
        if (err instanceof HttpError) {
            if (err.status === 409)
                submitError.value = t('errAlreadySubmitted');
            else if (err.status === 400)
                submitError.value = err.message || t('errInvalidBallot');
            else
                submitError.value = err.message;
        }
        else {
            submitError.value = err instanceof Error ? err.message : 'Network error';
        }
    }
    finally {
        submitting.value = false;
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
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_9 = {}.NText;
    /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
        tag: "h2",
        ...{ style: {} },
    }));
    const __VLS_11 = __VLS_10({
        tag: "h2",
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_10));
    __VLS_12.slots.default;
    (__VLS_ctx.context.gathering.title);
    var __VLS_12;
    if (__VLS_ctx.context.gathering.description) {
        const __VLS_13 = {}.NText;
        /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
        // @ts-ignore
        const __VLS_14 = __VLS_asFunctionalComponent(__VLS_13, new __VLS_13({
            depth: (2),
            ...{ style: {} },
        }));
        const __VLS_15 = __VLS_14({
            depth: (2),
            ...{ style: {} },
        }, ...__VLS_functionalComponentArgsRest(__VLS_14));
        __VLS_16.slots.default;
        (__VLS_ctx.context.gathering.description);
        var __VLS_16;
    }
    const __VLS_17 = {}.NCard;
    /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({
        ...{ style: {} },
    }));
    const __VLS_19 = __VLS_18({
        ...{ style: {} },
    }, ...__VLS_functionalComponentArgsRest(__VLS_18));
    __VLS_20.slots.default;
    const __VLS_21 = {}.NDescriptions;
    /** @type {[typeof __VLS_components.NDescriptions, typeof __VLS_components.NDescriptions, ]} */ ;
    // @ts-ignore
    const __VLS_22 = __VLS_asFunctionalComponent(__VLS_21, new __VLS_21({
        column: (3),
        labelPlacement: "top",
        size: "small",
    }));
    const __VLS_23 = __VLS_22({
        column: (3),
        labelPlacement: "top",
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_22));
    __VLS_24.slots.default;
    const __VLS_25 = {}.NDescriptionsItem;
    /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_26 = __VLS_asFunctionalComponent(__VLS_25, new __VLS_25({
        label: (__VLS_ctx.t('owner')),
    }));
    const __VLS_27 = __VLS_26({
        label: (__VLS_ctx.t('owner')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_26));
    __VLS_28.slots.default;
    (__VLS_ctx.context.owner.name);
    var __VLS_28;
    const __VLS_29 = {}.NDescriptionsItem;
    /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_30 = __VLS_asFunctionalComponent(__VLS_29, new __VLS_29({
        label: (__VLS_ctx.t('units')),
    }));
    const __VLS_31 = __VLS_30({
        label: (__VLS_ctx.t('units')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_30));
    __VLS_32.slots.default;
    (__VLS_ctx.context.units.length);
    var __VLS_32;
    const __VLS_33 = {}.NDescriptionsItem;
    /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
    // @ts-ignore
    const __VLS_34 = __VLS_asFunctionalComponent(__VLS_33, new __VLS_33({
        label: (__VLS_ctx.context.gathering.voting_mode === 'by_weight' ? __VLS_ctx.t('votingWeight') : __VLS_ctx.t('votingUnits')),
    }));
    const __VLS_35 = __VLS_34({
        label: (__VLS_ctx.context.gathering.voting_mode === 'by_weight' ? __VLS_ctx.t('votingWeight') : __VLS_ctx.t('votingUnits')),
    }, ...__VLS_functionalComponentArgsRest(__VLS_34));
    __VLS_36.slots.default;
    (__VLS_ctx.votingShare.toFixed(2));
    var __VLS_36;
    var __VLS_24;
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ style: {} },
    });
    const __VLS_37 = {}.NTag;
    /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
    // @ts-ignore
    const __VLS_38 = __VLS_asFunctionalComponent(__VLS_37, new __VLS_37({
        type: (__VLS_ctx.statusTagType),
        size: "small",
    }));
    const __VLS_39 = __VLS_38({
        type: (__VLS_ctx.statusTagType),
        size: "small",
    }, ...__VLS_functionalComponentArgsRest(__VLS_38));
    __VLS_40.slots.default;
    (__VLS_ctx.context.gathering.status.toUpperCase());
    var __VLS_40;
    var __VLS_20;
    if (__VLS_ctx.context.gathering.status !== 'active') {
        const __VLS_41 = {}.NAlert;
        /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
        // @ts-ignore
        const __VLS_42 = __VLS_asFunctionalComponent(__VLS_41, new __VLS_41({
            type: (__VLS_ctx.context.gathering.status === 'tallied' ? 'info' : 'warning'),
        }));
        const __VLS_43 = __VLS_42({
            type: (__VLS_ctx.context.gathering.status === 'tallied' ? 'info' : 'warning'),
        }, ...__VLS_functionalComponentArgsRest(__VLS_42));
        __VLS_44.slots.default;
        if (['draft', 'scheduled'].includes(__VLS_ctx.context.gathering.status)) {
            (__VLS_ctx.t('statusNotStarted'));
        }
        else if (__VLS_ctx.context.gathering.status === 'tallied') {
            (__VLS_ctx.t('statusTallied'));
        }
        else {
            (__VLS_ctx.t('statusClosed'));
        }
        var __VLS_44;
    }
    else {
        if (__VLS_ctx.allUnitsUnavailable && !__VLS_ctx.receipt) {
            const __VLS_45 = {}.NAlert;
            /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
            // @ts-ignore
            const __VLS_46 = __VLS_asFunctionalComponent(__VLS_45, new __VLS_45({
                type: "warning",
                ...{ style: {} },
            }));
            const __VLS_47 = __VLS_46({
                type: "warning",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_46));
            __VLS_48.slots.default;
            (__VLS_ctx.t('unitsAlreadyVoted'));
            var __VLS_48;
        }
        else if (__VLS_ctx.receipt) {
            const __VLS_49 = {}.NCard;
            /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
            // @ts-ignore
            const __VLS_50 = __VLS_asFunctionalComponent(__VLS_49, new __VLS_49({}));
            const __VLS_51 = __VLS_50({}, ...__VLS_functionalComponentArgsRest(__VLS_50));
            __VLS_52.slots.default;
            {
                const { header: __VLS_thisSlot } = __VLS_52.slots;
                __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({
                    ...{ style: {} },
                });
                (__VLS_ctx.t('ballotSubmitted'));
            }
            const __VLS_53 = {}.NDescriptions;
            /** @type {[typeof __VLS_components.NDescriptions, typeof __VLS_components.NDescriptions, ]} */ ;
            // @ts-ignore
            const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
                column: (1),
                labelPlacement: "left",
                size: "small",
                ...{ style: {} },
            }));
            const __VLS_55 = __VLS_54({
                column: (1),
                labelPlacement: "left",
                size: "small",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_54));
            __VLS_56.slots.default;
            const __VLS_57 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
                label: (__VLS_ctx.t('ballotId')),
            }));
            const __VLS_59 = __VLS_58({
                label: (__VLS_ctx.t('ballotId')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_58));
            __VLS_60.slots.default;
            (__VLS_ctx.receipt.ballot_id);
            var __VLS_60;
            const __VLS_61 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
                label: (__VLS_ctx.t('verificationHash')),
            }));
            const __VLS_63 = __VLS_62({
                label: (__VLS_ctx.t('verificationHash')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_62));
            __VLS_64.slots.default;
            const __VLS_65 = {}.NText;
            /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
            // @ts-ignore
            const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
                code: true,
                ...{ style: {} },
            }));
            const __VLS_67 = __VLS_66({
                code: true,
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_66));
            __VLS_68.slots.default;
            (__VLS_ctx.receipt.ballot_hash);
            var __VLS_68;
            var __VLS_64;
            const __VLS_69 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
                label: (__VLS_ctx.t('submittedAt')),
            }));
            const __VLS_71 = __VLS_70({
                label: (__VLS_ctx.t('submittedAt')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_70));
            __VLS_72.slots.default;
            (__VLS_ctx.receipt.submitted_at ? new Date(__VLS_ctx.receipt.submitted_at).toLocaleString() : '—');
            var __VLS_72;
            var __VLS_56;
            const __VLS_73 = {}.NDivider;
            /** @type {[typeof __VLS_components.NDivider, typeof __VLS_components.NDivider, ]} */ ;
            // @ts-ignore
            const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
                titlePlacement: "left",
            }));
            const __VLS_75 = __VLS_74({
                titlePlacement: "left",
            }, ...__VLS_functionalComponentArgsRest(__VLS_74));
            __VLS_76.slots.default;
            (__VLS_ctx.t('yourVotes'));
            var __VLS_76;
            for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.context.matters.filter(m => !m.is_informative)))) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                    key: (matter.id),
                    ...{ style: {} },
                });
                const __VLS_77 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
                    strong: true,
                    ...{ style: {} },
                }));
                const __VLS_79 = __VLS_78({
                    strong: true,
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_78));
                __VLS_80.slots.default;
                (__VLS_ctx.matterTitle(matter));
                var __VLS_80;
                const __VLS_81 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
                    depth: (2),
                    ...{ style: {} },
                }));
                const __VLS_83 = __VLS_82({
                    depth: (2),
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_82));
                __VLS_84.slots.default;
                (__VLS_ctx.formatVotedValues(matter));
                var __VLS_84;
            }
            var __VLS_52;
        }
        else {
            const __VLS_85 = {}.NCard;
            /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
            // @ts-ignore
            const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({}));
            const __VLS_87 = __VLS_86({}, ...__VLS_functionalComponentArgsRest(__VLS_86));
            __VLS_88.slots.default;
            {
                const { header: __VLS_thisSlot } = __VLS_88.slots;
                __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
                (__VLS_ctx.t('yourBallot'));
                const __VLS_89 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({
                    depth: (3),
                    ...{ style: {} },
                }));
                const __VLS_91 = __VLS_90({
                    depth: (3),
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_90));
                __VLS_92.slots.default;
                (__VLS_ctx.context.gathering.title);
                var __VLS_92;
            }
            if (__VLS_ctx.submitError) {
                const __VLS_93 = {}.NAlert;
                /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
                // @ts-ignore
                const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
                    ...{ 'onClose': {} },
                    type: "error",
                    closable: true,
                    ...{ style: {} },
                }));
                const __VLS_95 = __VLS_94({
                    ...{ 'onClose': {} },
                    type: "error",
                    closable: true,
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_94));
                let __VLS_97;
                let __VLS_98;
                let __VLS_99;
                const __VLS_100 = {
                    onClose: (...[$event]) => {
                        if (!(__VLS_ctx.context))
                            return;
                        if (!!(__VLS_ctx.context.gathering.status !== 'active'))
                            return;
                        if (!!(__VLS_ctx.allUnitsUnavailable && !__VLS_ctx.receipt))
                            return;
                        if (!!(__VLS_ctx.receipt))
                            return;
                        if (!(__VLS_ctx.submitError))
                            return;
                        __VLS_ctx.submitError = null;
                    }
                };
                __VLS_96.slots.default;
                (__VLS_ctx.submitError);
                var __VLS_96;
            }
            for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.informativeMatters))) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                    key: (matter.id),
                    ...{ style: {} },
                });
                const __VLS_101 = {}.NCard;
                /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
                // @ts-ignore
                const __VLS_102 = __VLS_asFunctionalComponent(__VLS_101, new __VLS_101({
                    size: "small",
                    embedded: true,
                }));
                const __VLS_103 = __VLS_102({
                    size: "small",
                    embedded: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_102));
                __VLS_104.slots.default;
                {
                    const { header: __VLS_thisSlot } = __VLS_104.slots;
                    const __VLS_105 = {}.NText;
                    /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                    // @ts-ignore
                    const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
                        ...{ style: {} },
                    }));
                    const __VLS_107 = __VLS_106({
                        ...{ style: {} },
                    }, ...__VLS_functionalComponentArgsRest(__VLS_106));
                    __VLS_108.slots.default;
                    (__VLS_ctx.matterTitle(matter));
                    var __VLS_108;
                    const __VLS_109 = {}.NTag;
                    /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
                    // @ts-ignore
                    const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
                        size: "tiny",
                        ...{ style: {} },
                    }));
                    const __VLS_111 = __VLS_110({
                        size: "tiny",
                        ...{ style: {} },
                    }, ...__VLS_functionalComponentArgsRest(__VLS_110));
                    __VLS_112.slots.default;
                    (__VLS_ctx.t('informational'));
                    var __VLS_112;
                }
                const __VLS_113 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
                    depth: (2),
                    ...{ style: {} },
                }));
                const __VLS_115 = __VLS_114({
                    depth: (2),
                    ...{ style: {} },
                }, ...__VLS_functionalComponentArgsRest(__VLS_114));
                __VLS_116.slots.default;
                (__VLS_ctx.matterDescription(matter));
                var __VLS_116;
                var __VLS_104;
            }
            /** @type {[typeof BallotForm, ]} */ ;
            // @ts-ignore
            const __VLS_117 = __VLS_asFunctionalComponent(BallotForm, new BallotForm({
                matters: (__VLS_ctx.votableMatters),
                modelValue: (__VLS_ctx.ballotVotes),
            }));
            const __VLS_118 = __VLS_117({
                matters: (__VLS_ctx.votableMatters),
                modelValue: (__VLS_ctx.ballotVotes),
            }, ...__VLS_functionalComponentArgsRest(__VLS_117));
            const __VLS_120 = {}.NSpace;
            /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
            // @ts-ignore
            const __VLS_121 = __VLS_asFunctionalComponent(__VLS_120, new __VLS_120({
                justify: "end",
                ...{ style: {} },
            }));
            const __VLS_122 = __VLS_121({
                justify: "end",
                ...{ style: {} },
            }, ...__VLS_functionalComponentArgsRest(__VLS_121));
            __VLS_123.slots.default;
            const __VLS_124 = {}.NButton;
            /** @type {[typeof __VLS_components.NButton, typeof __VLS_components.NButton, ]} */ ;
            // @ts-ignore
            const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.submitting),
                disabled: (!__VLS_ctx.canSubmit),
            }));
            const __VLS_126 = __VLS_125({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.submitting),
                disabled: (!__VLS_ctx.canSubmit),
            }, ...__VLS_functionalComponentArgsRest(__VLS_125));
            let __VLS_128;
            let __VLS_129;
            let __VLS_130;
            const __VLS_131 = {
                onClick: (__VLS_ctx.handleSubmit)
            };
            __VLS_127.slots.default;
            (__VLS_ctx.t('submitBallot'));
            var __VLS_127;
            var __VLS_123;
            var __VLS_88;
        }
    }
}
var __VLS_3;
var __VLS_dollars;
const __VLS_self = (await import('vue')).defineComponent({
    setup() {
        return {
            NAlert: NAlert,
            NButton: NButton,
            NCard: NCard,
            NDescriptions: NDescriptions,
            NDescriptionsItem: NDescriptionsItem,
            NDivider: NDivider,
            NSpace: NSpace,
            NSpin: NSpin,
            NTag: NTag,
            NText: NText,
            BallotForm: BallotForm,
            t: t,
            loading: loading,
            submitting: submitting,
            fetchError: fetchError,
            submitError: submitError,
            context: context,
            receipt: receipt,
            ballotVotes: ballotVotes,
            votingShare: votingShare,
            allUnitsUnavailable: allUnitsUnavailable,
            votableMatters: votableMatters,
            informativeMatters: informativeMatters,
            canSubmit: canSubmit,
            statusTagType: statusTagType,
            matterTitle: matterTitle,
            matterDescription: matterDescription,
            formatVotedValues: formatVotedValues,
            handleSubmit: handleSubmit,
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
