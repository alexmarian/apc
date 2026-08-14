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
/** @type {__VLS_StyleScopedClasses['vw-subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-header-note']} */ ;
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
        ...{ class: "vw-block" },
    }));
    const __VLS_7 = __VLS_6({
        type: "error",
        ...{ class: "vw-block" },
    }, ...__VLS_functionalComponentArgsRest(__VLS_6));
    __VLS_8.slots.default;
    (__VLS_ctx.fetchError);
    var __VLS_8;
}
if (__VLS_ctx.context) {
    __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
        ...{ class: "vw-block" },
    });
    const __VLS_9 = {}.NText;
    /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
    // @ts-ignore
    const __VLS_10 = __VLS_asFunctionalComponent(__VLS_9, new __VLS_9({
        tag: "h2",
        ...{ class: "vw-title" },
    }));
    const __VLS_11 = __VLS_10({
        tag: "h2",
        ...{ class: "vw-title" },
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
            ...{ class: "vw-subtitle" },
        }));
        const __VLS_15 = __VLS_14({
            depth: (2),
            ...{ class: "vw-subtitle" },
        }, ...__VLS_functionalComponentArgsRest(__VLS_14));
        __VLS_16.slots.default;
        (__VLS_ctx.context.gathering.description);
        var __VLS_16;
    }
    const __VLS_17 = {}.NCard;
    /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
    // @ts-ignore
    const __VLS_18 = __VLS_asFunctionalComponent(__VLS_17, new __VLS_17({
        ...{ class: "vw-block" },
    }));
    const __VLS_19 = __VLS_18({
        ...{ class: "vw-block" },
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
        ...{ class: "vw-status" },
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
                ...{ class: "vw-block" },
            }));
            const __VLS_47 = __VLS_46({
                type: "warning",
                ...{ class: "vw-block" },
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
                const __VLS_53 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_54 = __VLS_asFunctionalComponent(__VLS_53, new __VLS_53({
                    type: "success",
                }));
                const __VLS_55 = __VLS_54({
                    type: "success",
                }, ...__VLS_functionalComponentArgsRest(__VLS_54));
                __VLS_56.slots.default;
                (__VLS_ctx.t('ballotSubmitted'));
                var __VLS_56;
            }
            const __VLS_57 = {}.NDescriptions;
            /** @type {[typeof __VLS_components.NDescriptions, typeof __VLS_components.NDescriptions, ]} */ ;
            // @ts-ignore
            const __VLS_58 = __VLS_asFunctionalComponent(__VLS_57, new __VLS_57({
                column: (1),
                labelPlacement: "left",
                size: "small",
                ...{ class: "vw-block" },
            }));
            const __VLS_59 = __VLS_58({
                column: (1),
                labelPlacement: "left",
                size: "small",
                ...{ class: "vw-block" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_58));
            __VLS_60.slots.default;
            const __VLS_61 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_62 = __VLS_asFunctionalComponent(__VLS_61, new __VLS_61({
                label: (__VLS_ctx.t('ballotId')),
            }));
            const __VLS_63 = __VLS_62({
                label: (__VLS_ctx.t('ballotId')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_62));
            __VLS_64.slots.default;
            (__VLS_ctx.receipt.ballot_id);
            var __VLS_64;
            const __VLS_65 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_66 = __VLS_asFunctionalComponent(__VLS_65, new __VLS_65({
                label: (__VLS_ctx.t('verificationHash')),
            }));
            const __VLS_67 = __VLS_66({
                label: (__VLS_ctx.t('verificationHash')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_66));
            __VLS_68.slots.default;
            const __VLS_69 = {}.NText;
            /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
            // @ts-ignore
            const __VLS_70 = __VLS_asFunctionalComponent(__VLS_69, new __VLS_69({
                code: true,
                ...{ class: "vw-hash" },
            }));
            const __VLS_71 = __VLS_70({
                code: true,
                ...{ class: "vw-hash" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_70));
            __VLS_72.slots.default;
            (__VLS_ctx.receipt.ballot_hash);
            var __VLS_72;
            var __VLS_68;
            const __VLS_73 = {}.NDescriptionsItem;
            /** @type {[typeof __VLS_components.NDescriptionsItem, typeof __VLS_components.NDescriptionsItem, ]} */ ;
            // @ts-ignore
            const __VLS_74 = __VLS_asFunctionalComponent(__VLS_73, new __VLS_73({
                label: (__VLS_ctx.t('submittedAt')),
            }));
            const __VLS_75 = __VLS_74({
                label: (__VLS_ctx.t('submittedAt')),
            }, ...__VLS_functionalComponentArgsRest(__VLS_74));
            __VLS_76.slots.default;
            (__VLS_ctx.receipt.submitted_at ? new Date(__VLS_ctx.receipt.submitted_at).toLocaleString() : '—');
            var __VLS_76;
            var __VLS_60;
            const __VLS_77 = {}.NDivider;
            /** @type {[typeof __VLS_components.NDivider, typeof __VLS_components.NDivider, ]} */ ;
            // @ts-ignore
            const __VLS_78 = __VLS_asFunctionalComponent(__VLS_77, new __VLS_77({
                titlePlacement: "left",
            }));
            const __VLS_79 = __VLS_78({
                titlePlacement: "left",
            }, ...__VLS_functionalComponentArgsRest(__VLS_78));
            __VLS_80.slots.default;
            (__VLS_ctx.t('yourVotes'));
            var __VLS_80;
            for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.context.matters.filter(m => !m.is_informative)))) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                    key: (matter.id),
                    ...{ class: "vw-voted" },
                });
                const __VLS_81 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_82 = __VLS_asFunctionalComponent(__VLS_81, new __VLS_81({
                    strong: true,
                    ...{ class: "vw-voted__matter" },
                }));
                const __VLS_83 = __VLS_82({
                    strong: true,
                    ...{ class: "vw-voted__matter" },
                }, ...__VLS_functionalComponentArgsRest(__VLS_82));
                __VLS_84.slots.default;
                (__VLS_ctx.matterTitle(matter));
                var __VLS_84;
                const __VLS_85 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_86 = __VLS_asFunctionalComponent(__VLS_85, new __VLS_85({
                    depth: (2),
                    ...{ class: "vw-voted__values" },
                }));
                const __VLS_87 = __VLS_86({
                    depth: (2),
                    ...{ class: "vw-voted__values" },
                }, ...__VLS_functionalComponentArgsRest(__VLS_86));
                __VLS_88.slots.default;
                (__VLS_ctx.formatVotedValues(matter));
                var __VLS_88;
            }
            var __VLS_52;
        }
        else {
            const __VLS_89 = {}.NCard;
            /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
            // @ts-ignore
            const __VLS_90 = __VLS_asFunctionalComponent(__VLS_89, new __VLS_89({}));
            const __VLS_91 = __VLS_90({}, ...__VLS_functionalComponentArgsRest(__VLS_90));
            __VLS_92.slots.default;
            {
                const { header: __VLS_thisSlot } = __VLS_92.slots;
                __VLS_asFunctionalElement(__VLS_intrinsicElements.span, __VLS_intrinsicElements.span)({});
                (__VLS_ctx.t('yourBallot'));
                const __VLS_93 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_94 = __VLS_asFunctionalComponent(__VLS_93, new __VLS_93({
                    depth: (3),
                    ...{ class: "vw-header-note" },
                }));
                const __VLS_95 = __VLS_94({
                    depth: (3),
                    ...{ class: "vw-header-note" },
                }, ...__VLS_functionalComponentArgsRest(__VLS_94));
                __VLS_96.slots.default;
                (__VLS_ctx.context.gathering.title);
                var __VLS_96;
            }
            if (__VLS_ctx.submitError) {
                const __VLS_97 = {}.NAlert;
                /** @type {[typeof __VLS_components.NAlert, typeof __VLS_components.NAlert, ]} */ ;
                // @ts-ignore
                const __VLS_98 = __VLS_asFunctionalComponent(__VLS_97, new __VLS_97({
                    ...{ 'onClose': {} },
                    type: "error",
                    closable: true,
                    ...{ class: "vw-block" },
                }));
                const __VLS_99 = __VLS_98({
                    ...{ 'onClose': {} },
                    type: "error",
                    closable: true,
                    ...{ class: "vw-block" },
                }, ...__VLS_functionalComponentArgsRest(__VLS_98));
                let __VLS_101;
                let __VLS_102;
                let __VLS_103;
                const __VLS_104 = {
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
                __VLS_100.slots.default;
                (__VLS_ctx.submitError);
                var __VLS_100;
            }
            for (const [matter] of __VLS_getVForSourceType((__VLS_ctx.informativeMatters))) {
                __VLS_asFunctionalElement(__VLS_intrinsicElements.div, __VLS_intrinsicElements.div)({
                    key: (matter.id),
                    ...{ class: "vw-block" },
                });
                const __VLS_105 = {}.NCard;
                /** @type {[typeof __VLS_components.NCard, typeof __VLS_components.NCard, ]} */ ;
                // @ts-ignore
                const __VLS_106 = __VLS_asFunctionalComponent(__VLS_105, new __VLS_105({
                    size: "small",
                    embedded: true,
                }));
                const __VLS_107 = __VLS_106({
                    size: "small",
                    embedded: true,
                }, ...__VLS_functionalComponentArgsRest(__VLS_106));
                __VLS_108.slots.default;
                {
                    const { header: __VLS_thisSlot } = __VLS_108.slots;
                    const __VLS_109 = {}.NText;
                    /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                    // @ts-ignore
                    const __VLS_110 = __VLS_asFunctionalComponent(__VLS_109, new __VLS_109({
                        ...{ class: "vw-body-text" },
                    }));
                    const __VLS_111 = __VLS_110({
                        ...{ class: "vw-body-text" },
                    }, ...__VLS_functionalComponentArgsRest(__VLS_110));
                    __VLS_112.slots.default;
                    (__VLS_ctx.matterTitle(matter));
                    var __VLS_112;
                    const __VLS_113 = {}.NTag;
                    /** @type {[typeof __VLS_components.NTag, typeof __VLS_components.NTag, ]} */ ;
                    // @ts-ignore
                    const __VLS_114 = __VLS_asFunctionalComponent(__VLS_113, new __VLS_113({
                        size: "tiny",
                        ...{ class: "vw-header-tag" },
                    }));
                    const __VLS_115 = __VLS_114({
                        size: "tiny",
                        ...{ class: "vw-header-tag" },
                    }, ...__VLS_functionalComponentArgsRest(__VLS_114));
                    __VLS_116.slots.default;
                    (__VLS_ctx.t('informational'));
                    var __VLS_116;
                }
                const __VLS_117 = {}.NText;
                /** @type {[typeof __VLS_components.NText, typeof __VLS_components.NText, ]} */ ;
                // @ts-ignore
                const __VLS_118 = __VLS_asFunctionalComponent(__VLS_117, new __VLS_117({
                    depth: (2),
                    ...{ class: "vw-body-text" },
                }));
                const __VLS_119 = __VLS_118({
                    depth: (2),
                    ...{ class: "vw-body-text" },
                }, ...__VLS_functionalComponentArgsRest(__VLS_118));
                __VLS_120.slots.default;
                (__VLS_ctx.matterDescription(matter));
                var __VLS_120;
                var __VLS_108;
            }
            /** @type {[typeof BallotForm, ]} */ ;
            // @ts-ignore
            const __VLS_121 = __VLS_asFunctionalComponent(BallotForm, new BallotForm({
                matters: (__VLS_ctx.votableMatters),
                modelValue: (__VLS_ctx.ballotVotes),
            }));
            const __VLS_122 = __VLS_121({
                matters: (__VLS_ctx.votableMatters),
                modelValue: (__VLS_ctx.ballotVotes),
            }, ...__VLS_functionalComponentArgsRest(__VLS_121));
            const __VLS_124 = {}.NSpace;
            /** @type {[typeof __VLS_components.NSpace, typeof __VLS_components.NSpace, ]} */ ;
            // @ts-ignore
            const __VLS_125 = __VLS_asFunctionalComponent(__VLS_124, new __VLS_124({
                justify: "end",
                ...{ class: "vw-actions" },
            }));
            const __VLS_126 = __VLS_125({
                justify: "end",
                ...{ class: "vw-actions" },
            }, ...__VLS_functionalComponentArgsRest(__VLS_125));
            __VLS_127.slots.default;
            const __VLS_128 = {}.NButton;
            /** @type {[typeof __VLS_components.NButton, typeof __VLS_components.NButton, ]} */ ;
            // @ts-ignore
            const __VLS_129 = __VLS_asFunctionalComponent(__VLS_128, new __VLS_128({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.submitting),
                disabled: (!__VLS_ctx.canSubmit),
            }));
            const __VLS_130 = __VLS_129({
                ...{ 'onClick': {} },
                type: "primary",
                loading: (__VLS_ctx.submitting),
                disabled: (!__VLS_ctx.canSubmit),
            }, ...__VLS_functionalComponentArgsRest(__VLS_129));
            let __VLS_132;
            let __VLS_133;
            let __VLS_134;
            const __VLS_135 = {
                onClick: (__VLS_ctx.handleSubmit)
            };
            __VLS_131.slots.default;
            (__VLS_ctx.t('submitBallot'));
            var __VLS_131;
            var __VLS_127;
            var __VLS_92;
        }
    }
}
var __VLS_3;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-title']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-subtitle']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-status']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-hash']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-voted']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-voted__matter']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-voted__values']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-header-note']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-block']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-body-text']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-header-tag']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-body-text']} */ ;
/** @type {__VLS_StyleScopedClasses['vw-actions']} */ ;
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
