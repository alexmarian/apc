import { defineComponent as K, watch as re, openBlock as r, createElementBlock as _, Fragment as V, renderList as U, createVNode as o, unref as e, withCtx as t, createBlock as h, createTextVNode as n, toDisplayString as a, createCommentVNode as B, ref as q, computed as I, onMounted as ae, createElementVNode as G } from "vue";
import { useI18n as X } from "vue-i18n";
import { NCard as D, NText as N, NRadioGroup as ee, NSpace as M, NRadio as W, NCheckboxGroup as de, NCheckbox as ce, NButton as J, NSpin as le, NAlert as P, NDescriptions as L, NDescriptionsItem as z, NTag as Y, NDivider as ie, NProgress as ve } from "naive-ui";
const fe = { key: 4 }, ge = /* @__PURE__ */ K({
  __name: "BallotForm",
  props: {
    matters: {},
    modelValue: {}
  },
  emits: ["update:modelValue"],
  setup(S, { emit: i }) {
    const { t: k, locale: b } = X({
      useScope: "local",
      messages: {
        en: {
          yes: "Yes",
          no: "No",
          abstain: "Abstain",
          rankingHint: "Rank options from most preferred (top) to least preferred (bottom)."
        },
        ro: {
          yes: "Da",
          no: "Nu",
          abstain: "Abținere",
          rankingHint: "Ordonați opțiunile de la cea mai preferată (sus) la cea mai puțin preferată (jos)."
        },
        ru: {
          yes: "Да",
          no: "Нет",
          abstain: "Воздержаться",
          rankingHint: "Упорядочьте варианты от наиболее предпочтительного (вверху) до наименее предпочтительного (внизу)."
        }
      }
    }), y = S, m = i;
    re(
      () => y.matters,
      (d) => {
        const s = { ...y.modelValue };
        let l = !1;
        for (const v of d) {
          const w = String(v.id);
          s[w] === void 0 && (s[w] = v.voting_config.type === "ranking" ? (v.voting_config.options ?? []).map((H) => H.id) : [], l = !0);
        }
        l && m("update:modelValue", s);
      },
      { immediate: !0 }
    );
    function x(d) {
      var s;
      return ((s = y.modelValue[String(d)]) == null ? void 0 : s[0]) ?? "";
    }
    function A(d, s) {
      m("update:modelValue", { ...y.modelValue, [String(d)]: s ? [s] : [] });
    }
    function p(d) {
      return y.modelValue[String(d)] ?? [];
    }
    function C(d, s) {
      m("update:modelValue", { ...y.modelValue, [String(d)]: s.map(String) });
    }
    function T(d) {
      return y.modelValue[String(d)] ?? [];
    }
    function E(d, s, l) {
      const v = [...T(d)], w = s + l;
      w < 0 || w >= v.length || ([v[s], v[w]] = [v[w], v[s]], m("update:modelValue", { ...y.modelValue, [String(d)]: v }));
    }
    function $(d) {
      var l;
      return ((l = b.value) == null ? void 0 : l.slice(0, 2)) === "ru" && d.title_ru ? d.title_ru : d.title;
    }
    function j(d) {
      var l;
      return ((l = b.value) == null ? void 0 : l.slice(0, 2)) === "ru" && d.description_ru ? d.description_ru : d.description;
    }
    function g(d, s) {
      var l, v;
      return ((v = (l = d.voting_config.options) == null ? void 0 : l.find((w) => w.id === s)) == null ? void 0 : v.text) ?? s;
    }
    return (d, s) => (r(), _("div", null, [
      (r(!0), _(V, null, U(S.matters, (l) => (r(), _("div", {
        key: l.id,
        class: "bf-matter"
      }, [
        o(e(D), { size: "small" }, {
          header: t(() => [
            n(a($(l)), 1)
          ]),
          default: t(() => [
            l.description || l.description_ru ? (r(), h(e(N), {
              key: 0,
              depth: 2,
              class: "bf-description"
            }, {
              default: t(() => [
                n(a(j(l)), 1)
              ]),
              _: 2
            }, 1024)) : B("", !0),
            l.voting_config.type === "yes_no" ? (r(), h(e(ee), {
              key: 1,
              value: x(l.id),
              "onUpdate:value": (v) => A(l.id, v)
            }, {
              default: t(() => [
                o(e(M), null, {
                  default: t(() => [
                    o(e(W), { value: "yes" }, {
                      default: t(() => [
                        n(a(e(k)("yes")), 1)
                      ]),
                      _: 1
                    }),
                    o(e(W), { value: "no" }, {
                      default: t(() => [
                        n(a(e(k)("no")), 1)
                      ]),
                      _: 1
                    }),
                    l.voting_config.allow_abstention ? (r(), h(e(W), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: t(() => [
                        n(a(e(k)("abstain")), 1)
                      ]),
                      _: 1
                    })) : B("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "single_choice" ? (r(), h(e(ee), {
              key: 2,
              value: x(l.id),
              "onUpdate:value": (v) => A(l.id, v)
            }, {
              default: t(() => [
                o(e(M), { vertical: "" }, {
                  default: t(() => [
                    (r(!0), _(V, null, U(l.voting_config.options ?? [], (v) => (r(), h(e(W), {
                      key: v.id,
                      value: v.id
                    }, {
                      default: t(() => [
                        n(a(v.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128)),
                    l.voting_config.allow_abstention ? (r(), h(e(W), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: t(() => [
                        n(a(e(k)("abstain")), 1)
                      ]),
                      _: 1
                    })) : B("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "multiple_choice" ? (r(), h(e(de), {
              key: 3,
              value: p(l.id),
              "onUpdate:value": (v) => C(l.id, v)
            }, {
              default: t(() => [
                o(e(M), { vertical: "" }, {
                  default: t(() => [
                    (r(!0), _(V, null, U(l.voting_config.options ?? [], (v) => (r(), h(e(ce), {
                      key: v.id,
                      value: v.id
                    }, {
                      default: t(() => [
                        n(a(v.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128))
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "ranking" ? (r(), _("div", fe, [
              o(e(N), {
                depth: 3,
                class: "bf-hint"
              }, {
                default: t(() => [
                  n(a(e(k)("rankingHint")), 1)
                ]),
                _: 1
              }),
              (r(!0), _(V, null, U(T(l.id), (v, w) => (r(), _("div", {
                key: v,
                class: "bf-rank-row"
              }, [
                o(e(N), {
                  depth: 3,
                  class: "bf-rank-row__index"
                }, {
                  default: t(() => [
                    n(a(w + 1), 1)
                  ]),
                  _: 2
                }, 1024),
                o(e(N), { class: "bf-rank-row__label" }, {
                  default: t(() => [
                    n(a(g(l, v)), 1)
                  ]),
                  _: 2
                }, 1024),
                o(e(J), {
                  size: "tiny",
                  disabled: w === 0,
                  onClick: (H) => E(l.id, w, -1)
                }, {
                  default: t(() => [...s[0] || (s[0] = [
                    n("↑", -1)
                  ])]),
                  _: 1
                }, 8, ["disabled", "onClick"]),
                o(e(J), {
                  size: "tiny",
                  disabled: w === T(l.id).length - 1,
                  onClick: (H) => E(l.id, w, 1)
                }, {
                  default: t(() => [...s[1] || (s[1] = [
                    n("↓", -1)
                  ])]),
                  _: 1
                }, 8, ["disabled", "onClick"])
              ]))), 128))
            ])) : B("", !0)
          ]),
          _: 2
        }, 1024)
      ]))), 128))
    ]));
  }
}), Z = (S, i) => {
  const k = S.__vccOpts || S;
  for (const [b, y] of i)
    k[b] = y;
  return k;
}, pe = /* @__PURE__ */ Z(ge, [["__scopeId", "data-v-b97c74aa"]]);
class Q extends Error {
  constructor(i, k) {
    super(i), this.status = k, this.name = "HttpError";
  }
}
const _e = { class: "vw-block" }, be = { class: "vw-status" }, ye = /* @__PURE__ */ K({
  __name: "VotingWidget",
  props: {
    service: {},
    initialContext: {}
  },
  setup(S) {
    const { t: i, locale: k } = X({
      useScope: "local",
      messages: {
        en: {
          owner: "Owner",
          units: "Units",
          votingWeight: "Voting weight (%)",
          votingUnits: "Voting units",
          yourBallot: "Your Ballot",
          submitBallot: "Submit Ballot",
          ballotSubmitted: "Ballot Submitted",
          ballotId: "Ballot ID",
          verificationHash: "Verification hash",
          submittedAt: "Submitted at",
          yourVotes: "Your votes",
          informational: "Informational",
          yes: "Yes",
          no: "No",
          abstain: "Abstain",
          statusNotStarted: "Voting has not started yet. Please check back later.",
          statusTallied: "Voting has closed. Results are being tallied.",
          statusClosed: "Voting is closed.",
          errAlreadySubmitted: "A ballot has already been submitted for this gathering.",
          errInvalidBallot: "Invalid ballot.",
          unitsAlreadyVoted: "The voting rights for your unit(s) have already been exercised by another co-owner."
        },
        ro: {
          owner: "Proprietar",
          units: "Unități",
          votingWeight: "Pondere de vot (%)",
          votingUnits: "Unități de vot",
          yourBallot: "Buletinul dvs. de vot",
          submitBallot: "Trimite buletinul",
          ballotSubmitted: "Buletin trimis",
          ballotId: "ID buletin",
          verificationHash: "Hash de verificare",
          submittedAt: "Trimis la",
          yourVotes: "Voturile dvs.",
          informational: "Informativ",
          yes: "Da",
          no: "Nu",
          abstain: "Abținere",
          statusNotStarted: "Votul nu a început încă. Verificați mai târziu.",
          statusTallied: "Votul s-a încheiat. Rezultatele sunt în curs de numărare.",
          statusClosed: "Votul este închis.",
          errAlreadySubmitted: "Un buletin de vot a fost deja trimis pentru această adunare.",
          errInvalidBallot: "Buletin de vot invalid.",
          unitsAlreadyVoted: "Dreptul de vot pentru unitatea dvs. a fost deja exercitat de un alt coproprietar."
        },
        ru: {
          owner: "Владелец",
          units: "Единицы",
          votingWeight: "Вес голоса (%)",
          votingUnits: "Единицы голосования",
          yourBallot: "Ваш бюллетень",
          submitBallot: "Подать бюллетень",
          ballotSubmitted: "Бюллетень подан",
          ballotId: "ID бюллетеня",
          verificationHash: "Хэш верификации",
          submittedAt: "Подан в",
          yourVotes: "Ваши голоса",
          informational: "Информационный",
          yes: "Да",
          no: "Нет",
          abstain: "Воздержаться",
          statusNotStarted: "Голосование ещё не началось. Загляните позже.",
          statusTallied: "Голосование завершено. Результаты подсчитываются.",
          statusClosed: "Голосование закрыто.",
          errAlreadySubmitted: "Бюллетень для этого собрания уже был подан.",
          errInvalidBallot: "Недействительный бюллетень.",
          unitsAlreadyVoted: "Право голоса за вашу единицу уже было реализовано другим совладельцем."
        }
      }
    }), b = S, y = q(!1), m = q(!1), x = q(null), A = q(null), p = q(null), C = q(null), T = q({}), E = I(
      () => {
        var u;
        return ((u = p.value) == null ? void 0 : u.units.reduce((c, f) => c + f.voting_weight, 0)) ?? 0;
      }
    ), $ = I(() => {
      var c, f;
      const u = (c = p.value) == null ? void 0 : c.gathering;
      if (!u) return 0;
      if (u.voting_mode === "by_weight") {
        const R = u.qualified_units_total_part;
        return R > 0 ? E.value / R * 100 : 0;
      } else {
        const R = u.qualified_units_count;
        return R > 0 ? (((f = p.value) == null ? void 0 : f.units.length) ?? 0) / R * 100 : 0;
      }
    }), j = I(
      () => {
        var u;
        return (((u = p.value) == null ? void 0 : u.units.length) ?? 0) > 0 && p.value.units.every((c) => !c.is_available);
      }
    ), g = I(
      () => {
        var u;
        return (((u = p.value) == null ? void 0 : u.matters) ?? []).filter((c) => !c.is_informative);
      }
    ), d = I(
      () => {
        var u;
        return (((u = p.value) == null ? void 0 : u.matters) ?? []).filter((c) => c.is_informative);
      }
    ), s = I(
      () => g.value.length > 0 && g.value.every((u) => {
        var c;
        return (((c = T.value[String(u.id)]) == null ? void 0 : c.length) ?? 0) > 0;
      })
    ), l = I(() => {
      var u;
      switch ((u = p.value) == null ? void 0 : u.gathering.status) {
        case "active":
          return "success";
        case "scheduled":
          return "info";
        case "tallied":
          return "info";
        case "closed":
          return "error";
        default:
          return "default";
      }
    });
    function v(u) {
      var f;
      return ((f = k.value) == null ? void 0 : f.slice(0, 2)) === "ru" && u.title_ru ? u.title_ru : u.title;
    }
    function w(u) {
      var f;
      return ((f = k.value) == null ? void 0 : f.slice(0, 2)) === "ru" && u.description_ru ? u.description_ru : u.description;
    }
    function H() {
      const u = {};
      for (const c of g.value) {
        const f = String(c.id);
        u[f] = { matter_id: c.id, values: T.value[f] ?? [] };
      }
      return u;
    }
    function ne(u) {
      if (!C.value) return "—";
      const c = C.value.ballot_content[String(u.id)];
      return !c || c.values.length === 0 ? "—" : u.voting_config.type === "ranking" ? c.values.map((f, R) => {
        var O;
        const F = (O = u.voting_config.options) == null ? void 0 : O.find((ue) => ue.id === f);
        return `${R + 1}. ${F ? F.text : f}`;
      }).join(", ") : c.values.map((f) => {
        var F;
        if (f === "abstain") return i("abstain");
        if (u.voting_config.type === "yes_no") return i(f === "yes" ? "yes" : "no");
        const R = (F = u.voting_config.options) == null ? void 0 : F.find((O) => O.id === f);
        return R ? R.text : f;
      }).join(", ");
    }
    async function oe() {
      if (b.initialContext) {
        p.value = b.initialContext, b.initialContext.ballot && (C.value = {
          ballot_id: b.initialContext.ballot.ballot_id,
          ballot_hash: b.initialContext.ballot.ballot_hash,
          submitted_at: b.initialContext.ballot.submitted_at,
          ballot_content: b.initialContext.ballot.ballot_content
        });
        return;
      }
      y.value = !0, x.value = null;
      try {
        const u = await b.service.getContext();
        p.value = u, u.ballot && (C.value = {
          ballot_id: u.ballot.ballot_id,
          ballot_hash: u.ballot.ballot_hash,
          submitted_at: u.ballot.submitted_at,
          ballot_content: u.ballot.ballot_content
        });
      } catch (u) {
        x.value = u instanceof Error ? u.message : "Network error";
      } finally {
        y.value = !1;
      }
    }
    async function se() {
      if (!s.value) return;
      m.value = !0, A.value = null;
      const u = H();
      try {
        const c = await b.service.submitBallot(u);
        C.value = {
          ballot_id: c.ballot_id,
          ballot_hash: c.ballot_hash,
          submitted_at: c.submitted_at,
          ballot_content: c.ballot_content ?? u
        };
      } catch (c) {
        c instanceof Q ? c.status === 409 ? A.value = i("errAlreadySubmitted") : c.status === 400 ? A.value = c.message || i("errInvalidBallot") : A.value = c.message : A.value = c instanceof Error ? c.message : "Network error";
      } finally {
        m.value = !1;
      }
    }
    return ae(oe), (u, c) => (r(), h(e(le), { show: y.value }, {
      default: t(() => [
        x.value ? (r(), h(e(P), {
          key: 0,
          type: "error",
          class: "vw-block"
        }, {
          default: t(() => [
            n(a(x.value), 1)
          ]),
          _: 1
        })) : B("", !0),
        p.value ? (r(), _(V, { key: 1 }, [
          G("div", _e, [
            o(e(N), {
              tag: "h2",
              class: "vw-title"
            }, {
              default: t(() => [
                n(a(p.value.gathering.title), 1)
              ]),
              _: 1
            }),
            p.value.gathering.description ? (r(), h(e(N), {
              key: 0,
              depth: 2,
              class: "vw-subtitle"
            }, {
              default: t(() => [
                n(a(p.value.gathering.description), 1)
              ]),
              _: 1
            })) : B("", !0)
          ]),
          o(e(D), { class: "vw-block" }, {
            default: t(() => [
              o(e(L), {
                column: 3,
                "label-placement": "top",
                size: "small"
              }, {
                default: t(() => [
                  o(e(z), {
                    label: e(i)("owner")
                  }, {
                    default: t(() => [
                      n(a(p.value.owner.name), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(e(z), {
                    label: e(i)("units")
                  }, {
                    default: t(() => [
                      n(a(p.value.units.length), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(e(z), {
                    label: p.value.gathering.voting_mode === "by_weight" ? e(i)("votingWeight") : e(i)("votingUnits")
                  }, {
                    default: t(() => [
                      n(a($.value.toFixed(2)) + "% ", 1)
                    ]),
                    _: 1
                  }, 8, ["label"])
                ]),
                _: 1
              }),
              G("div", be, [
                o(e(Y), {
                  type: l.value,
                  size: "small"
                }, {
                  default: t(() => [
                    n(a(p.value.gathering.status.toUpperCase()), 1)
                  ]),
                  _: 1
                }, 8, ["type"])
              ])
            ]),
            _: 1
          }),
          p.value.gathering.status !== "active" ? (r(), h(e(P), {
            key: 0,
            type: p.value.gathering.status === "tallied" ? "info" : "warning"
          }, {
            default: t(() => [
              ["draft", "scheduled"].includes(p.value.gathering.status) ? (r(), _(V, { key: 0 }, [
                n(a(e(i)("statusNotStarted")), 1)
              ], 64)) : p.value.gathering.status === "tallied" ? (r(), _(V, { key: 1 }, [
                n(a(e(i)("statusTallied")), 1)
              ], 64)) : (r(), _(V, { key: 2 }, [
                n(a(e(i)("statusClosed")), 1)
              ], 64))
            ]),
            _: 1
          }, 8, ["type"])) : (r(), _(V, { key: 1 }, [
            j.value && !C.value ? (r(), h(e(P), {
              key: 0,
              type: "warning",
              class: "vw-block"
            }, {
              default: t(() => [
                n(a(e(i)("unitsAlreadyVoted")), 1)
              ]),
              _: 1
            })) : C.value ? (r(), h(e(D), { key: 1 }, {
              header: t(() => [
                o(e(N), { type: "success" }, {
                  default: t(() => [
                    n("✓ " + a(e(i)("ballotSubmitted")), 1)
                  ]),
                  _: 1
                })
              ]),
              default: t(() => [
                o(e(L), {
                  column: 1,
                  "label-placement": "left",
                  size: "small",
                  class: "vw-block"
                }, {
                  default: t(() => [
                    o(e(z), {
                      label: e(i)("ballotId")
                    }, {
                      default: t(() => [
                        n(a(C.value.ballot_id), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(e(z), {
                      label: e(i)("verificationHash")
                    }, {
                      default: t(() => [
                        o(e(N), {
                          code: "",
                          class: "vw-hash"
                        }, {
                          default: t(() => [
                            n(a(C.value.ballot_hash), 1)
                          ]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(e(z), {
                      label: e(i)("submittedAt")
                    }, {
                      default: t(() => [
                        n(a(C.value.submitted_at ? new Date(C.value.submitted_at).toLocaleString() : "—"), 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                }),
                o(e(ie), { "title-placement": "left" }, {
                  default: t(() => [
                    n(a(e(i)("yourVotes")), 1)
                  ]),
                  _: 1
                }),
                (r(!0), _(V, null, U(p.value.matters.filter((f) => !f.is_informative), (f) => (r(), _("div", {
                  key: f.id,
                  class: "vw-voted"
                }, [
                  o(e(N), {
                    strong: "",
                    class: "vw-voted__matter"
                  }, {
                    default: t(() => [
                      n(a(v(f)), 1)
                    ]),
                    _: 2
                  }, 1024),
                  o(e(N), {
                    depth: 2,
                    class: "vw-voted__values"
                  }, {
                    default: t(() => [
                      n(a(ne(f)), 1)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128))
              ]),
              _: 1
            })) : (r(), h(e(D), { key: 2 }, {
              header: t(() => [
                G("span", null, a(e(i)("yourBallot")), 1),
                o(e(N), {
                  depth: 3,
                  class: "vw-header-note"
                }, {
                  default: t(() => [
                    n(" — " + a(p.value.gathering.title), 1)
                  ]),
                  _: 1
                })
              ]),
              default: t(() => [
                A.value ? (r(), h(e(P), {
                  key: 0,
                  type: "error",
                  closable: "",
                  class: "vw-block",
                  onClose: c[0] || (c[0] = (f) => A.value = null)
                }, {
                  default: t(() => [
                    n(a(A.value), 1)
                  ]),
                  _: 1
                })) : B("", !0),
                (r(!0), _(V, null, U(d.value, (f) => (r(), _("div", {
                  key: f.id,
                  class: "vw-block"
                }, [
                  o(e(D), {
                    size: "small",
                    embedded: ""
                  }, {
                    header: t(() => [
                      o(e(N), { class: "vw-body-text" }, {
                        default: t(() => [
                          n(a(v(f)), 1)
                        ]),
                        _: 2
                      }, 1024),
                      o(e(Y), {
                        size: "tiny",
                        class: "vw-header-tag"
                      }, {
                        default: t(() => [
                          n(a(e(i)("informational")), 1)
                        ]),
                        _: 1
                      })
                    ]),
                    default: t(() => [
                      o(e(N), {
                        depth: 2,
                        class: "vw-body-text"
                      }, {
                        default: t(() => [
                          n(a(w(f)), 1)
                        ]),
                        _: 2
                      }, 1024)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128)),
                o(pe, {
                  matters: g.value,
                  modelValue: T.value,
                  "onUpdate:modelValue": c[1] || (c[1] = (f) => T.value = f)
                }, null, 8, ["matters", "modelValue"]),
                o(e(M), {
                  justify: "end",
                  class: "vw-actions"
                }, {
                  default: t(() => [
                    o(e(J), {
                      type: "primary",
                      loading: m.value,
                      disabled: !s.value,
                      onClick: se
                    }, {
                      default: t(() => [
                        n(a(e(i)("submitBallot")), 1)
                      ]),
                      _: 1
                    }, 8, ["loading", "disabled"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }))
          ], 64))
        ], 64)) : B("", !0)
      ]),
      _: 1
    }, 8, ["show"]));
  }
}), Ve = /* @__PURE__ */ Z(ye, [["__scopeId", "data-v-f0a6e1ae"]]), me = /* @__PURE__ */ K({
  __name: "VotingResultsWidget",
  props: {
    service: {},
    initialContext: {}
  },
  setup(S) {
    const { t: i } = X({
      useScope: "local",
      messages: {
        en: {
          gathering: "Gathering",
          status: "Status",
          participationSummary: "Participation Summary",
          participated: "Participated",
          voted: "Voted",
          units: "units",
          participationRate: "Participation rate",
          passed: "PASSED",
          failed: "FAILED",
          yourVoteCounted: "Your vote has been counted for this matter.",
          didNotVote: "You did not vote on this matter.",
          yourVote: "Your vote",
          vote: "vote",
          votes: "votes",
          yes: "Yes",
          no: "No",
          abstain: "Abstain",
          weight: "weight",
          quorum: "Quorum",
          quorumMet: "Met",
          quorumNotMet: "Not met",
          of: "of",
          required: "required",
          notAvailable: "Results are not yet available. Current status:",
          willBePublished: "Results will be published after the gathering is tallied."
        },
        ro: {
          gathering: "Adunare",
          status: "Status",
          participationSummary: "Rezumat participare",
          participated: "Participat",
          voted: "Votat",
          units: "unități",
          participationRate: "Rata de participare",
          passed: "ADOPTAT",
          failed: "RESPINS",
          yourVoteCounted: "Votul dvs. a fost înregistrat pentru acest punct.",
          didNotVote: "Nu ați votat pentru acest punct.",
          yourVote: "Votul dvs.",
          vote: "vot",
          votes: "voturi",
          yes: "Da",
          no: "Nu",
          abstain: "Abținere",
          weight: "pondere",
          quorum: "Cvorum",
          quorumMet: "Întrunit",
          quorumNotMet: "Neîntrunit",
          of: "din",
          required: "necesar",
          notAvailable: "Rezultatele nu sunt disponibile încă. Stare curentă:",
          willBePublished: "Rezultatele vor fi publicate după numărarea voturilor."
        },
        ru: {
          gathering: "Собрание",
          status: "Статус",
          participationSummary: "Сводка участия",
          participated: "Участвовало",
          voted: "Проголосовало",
          units: "ед.",
          participationRate: "Явка",
          passed: "ПРИНЯТО",
          failed: "ОТКЛОНЕНО",
          yourVoteCounted: "Ваш голос учтён по данному вопросу.",
          didNotVote: "Вы не голосовали по данному вопросу.",
          yourVote: "Ваш голос",
          vote: "голос",
          votes: "голосов",
          yes: "Да",
          no: "Нет",
          abstain: "Воздержаться",
          weight: "вес",
          quorum: "Кворум",
          quorumMet: "Достигнут",
          quorumNotMet: "Не достигнут",
          of: "из",
          required: "требуется",
          notAvailable: "Результаты пока недоступны. Текущий статус:",
          willBePublished: "Результаты будут опубликованы после подсчёта голосов."
        }
      }
    }), k = S, b = q(!1), y = q(null), m = q(null), x = q(null), A = I(() => {
      var g;
      switch ((g = m.value) == null ? void 0 : g.gathering.status) {
        case "active":
          return "success";
        case "scheduled":
          return "info";
        case "tallied":
          return "info";
        case "closed":
          return "error";
        default:
          return "default";
      }
    });
    function p(g) {
      var s;
      if (!((s = m.value) != null && s.ballot)) return !1;
      const d = m.value.ballot.ballot_content[String(g)];
      return !!d && d.values.length > 0;
    }
    function C(g, d) {
      var l;
      if (!((l = m.value) != null && l.ballot)) return !1;
      const s = m.value.ballot.ballot_content[String(g)];
      return !!s && s.values.includes(d);
    }
    function T(g, d) {
      var l;
      if (g === "abstain") return i("abstain");
      if (d.type === "yes_no") return i(g === "yes" ? "yes" : "no");
      const s = (l = d.options) == null ? void 0 : l.find((v) => v.id === g);
      return s ? s.text : g;
    }
    function E(g) {
      return [...g.votes].sort((d, s) => s.vote_count - d.vote_count);
    }
    function $(g, d) {
      if (g === "abstain") return "warning";
      if (d.voting_config.type === "yes_no") {
        if (g === "yes") return d.is_passed ? "success" : "default";
        if (g === "no") return d.is_passed ? "default" : "error";
      }
      return "default";
    }
    async function j() {
      if (k.initialContext) {
        m.value = k.initialContext, x.value = k.initialContext.results ?? null;
        return;
      }
      b.value = !0, y.value = null;
      try {
        const g = await k.service.getContext();
        m.value = g, x.value = g.results ?? null;
      } catch (g) {
        y.value = g instanceof Error ? g.message : "Network error";
      } finally {
        b.value = !1;
      }
    }
    return ae(j), (g, d) => (r(), h(e(le), { show: b.value }, {
      default: t(() => [
        y.value ? (r(), h(e(P), {
          key: 0,
          type: "error",
          class: "vr-block"
        }, {
          default: t(() => [
            n(a(y.value), 1)
          ]),
          _: 1
        })) : B("", !0),
        m.value ? (r(), _(V, { key: 1 }, [
          o(e(D), { class: "vr-block" }, {
            default: t(() => [
              o(e(L), {
                column: 2,
                "label-placement": "top",
                size: "small"
              }, {
                default: t(() => [
                  o(e(z), {
                    label: e(i)("gathering")
                  }, {
                    default: t(() => [
                      n(a(m.value.gathering.title), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(e(z), {
                    label: e(i)("status")
                  }, {
                    default: t(() => [
                      o(e(Y), {
                        type: A.value,
                        size: "small"
                      }, {
                        default: t(() => [
                          n(a(m.value.gathering.status.toUpperCase()), 1)
                        ]),
                        _: 1
                      }, 8, ["type"])
                    ]),
                    _: 1
                  }, 8, ["label"])
                ]),
                _: 1
              })
            ]),
            _: 1
          }),
          x.value ? (r(), _(V, { key: 1 }, [
            o(e(D), {
              size: "small",
              class: "vr-block",
              title: e(i)("participationSummary")
            }, {
              default: t(() => [
                o(e(L), {
                  column: 3,
                  "label-placement": "top",
                  size: "small"
                }, {
                  default: t(() => [
                    o(e(z), {
                      label: e(i)("participated")
                    }, {
                      default: t(() => [
                        n(a(x.value.statistics.participating_units) + " " + a(e(i)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(e(z), {
                      label: e(i)("voted")
                    }, {
                      default: t(() => [
                        n(a(x.value.statistics.voted_units) + " " + a(e(i)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(e(z), {
                      label: e(i)("participationRate")
                    }, {
                      default: t(() => [
                        n(a(x.value.statistics.participation_rate.toFixed(1)) + "% ", 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["title"]),
            (r(!0), _(V, null, U(x.value.results, (s) => (r(), _("div", {
              key: s.matter_id,
              class: "vr-block"
            }, [
              o(e(D), { size: "small" }, {
                header: t(() => [
                  o(e(M), {
                    align: "center",
                    justify: "space-between",
                    size: 4
                  }, {
                    default: t(() => [
                      o(e(N), { strong: "" }, {
                        default: t(() => [
                          n(a(s.matter_title), 1)
                        ]),
                        _: 2
                      }, 1024),
                      o(e(Y), {
                        type: s.is_passed ? "success" : "error",
                        size: "small"
                      }, {
                        default: t(() => [
                          n(a(s.is_passed ? e(i)("passed") : e(i)("failed")), 1)
                        ]),
                        _: 2
                      }, 1032, ["type"])
                    ]),
                    _: 2
                  }, 1024)
                ]),
                default: t(() => [
                  p(s.matter_id) ? (r(), h(e(P), {
                    key: 0,
                    type: "success",
                    size: "small",
                    class: "vr-vote-notice"
                  }, {
                    default: t(() => [
                      n(a(e(i)("yourVoteCounted")), 1)
                    ]),
                    _: 1
                  })) : m.value.ballot ? (r(), h(e(P), {
                    key: 1,
                    type: "default",
                    size: "small",
                    class: "vr-vote-notice"
                  }, {
                    default: t(() => [
                      n(a(e(i)("didNotVote")), 1)
                    ]),
                    _: 1
                  })) : B("", !0),
                  (r(!0), _(V, null, U(E(s), (l) => (r(), _("div", {
                    key: l.choice,
                    class: "vr-vote"
                  }, [
                    o(e(M), {
                      align: "center",
                      justify: "space-between",
                      class: "vr-vote__head"
                    }, {
                      default: t(() => [
                        o(e(M), {
                          align: "center",
                          size: "small"
                        }, {
                          default: t(() => [
                            o(e(N), {
                              type: C(s.matter_id, l.choice) ? "success" : "default",
                              strong: C(s.matter_id, l.choice)
                            }, {
                              default: t(() => [
                                n(a(T(l.choice, s.voting_config)), 1)
                              ]),
                              _: 2
                            }, 1032, ["type", "strong"]),
                            C(s.matter_id, l.choice) ? (r(), h(e(Y), {
                              key: 0,
                              type: "success",
                              size: "tiny"
                            }, {
                              default: t(() => [
                                n(a(e(i)("yourVote")), 1)
                              ]),
                              _: 1
                            })) : B("", !0)
                          ]),
                          _: 2
                        }, 1024),
                        o(e(N), {
                          depth: 2,
                          class: "vr-meta"
                        }, {
                          default: t(() => {
                            var v, w;
                            return [
                              n(a(l.vote_count) + " " + a(l.vote_count !== 1 ? e(i)("votes") : e(i)("vote")) + " (" + a(l.percentage.toFixed(1)) + "%", 1),
                              ((w = (v = x.value) == null ? void 0 : v.statistics) == null ? void 0 : w.voting_mode) === "by_weight" ? (r(), _(V, { key: 0 }, [
                                n(" · " + a(e(i)("weight")) + ": " + a(l.weight_percentage.toFixed(1)) + "%", 1)
                              ], 64)) : B("", !0),
                              d[1] || (d[1] = n(") ", -1))
                            ];
                          }),
                          _: 2
                        }, 1024)
                      ]),
                      _: 2
                    }, 1024),
                    o(e(ve), {
                      type: "line",
                      percentage: l.percentage,
                      status: $(l.choice, s),
                      "show-indicator": !1,
                      height: 8,
                      "border-radius": 4
                    }, null, 8, ["percentage", "status"])
                  ]))), 128)),
                  o(e(ie), { class: "vr-divider" }),
                  o(e(N), {
                    depth: 3,
                    class: "vr-meta"
                  }, {
                    default: t(() => {
                      var l;
                      return [
                        n(a(e(i)("quorum")) + ": " + a((l = s.quorum_info) != null && l.met ? e(i)("quorumMet") : e(i)("quorumNotMet")) + " ", 1),
                        s.quorum_info ? (r(), _(V, { key: 0 }, [
                          n(" — " + a(s.quorum_info.achieved_percentage.toFixed(1)) + "% " + a(e(i)("of")) + " " + a(s.quorum_info.required_percentage) + "% " + a(e(i)("required")), 1)
                        ], 64)) : B("", !0)
                      ];
                    }),
                    _: 2
                  }, 1024)
                ]),
                _: 2
              }, 1024)
            ]))), 128))
          ], 64)) : (r(), h(e(P), {
            key: 0,
            type: "info"
          }, {
            default: t(() => [
              n(a(e(i)("notAvailable")) + " ", 1),
              G("strong", null, a(m.value.gathering.status), 1),
              d[0] || (d[0] = n(". ", -1)),
              m.value.gathering.status !== "tallied" ? (r(), _(V, { key: 0 }, [
                n(a(e(i)("willBePublished")), 1)
              ], 64)) : B("", !0)
            ]),
            _: 1
          }))
        ], 64)) : B("", !0)
      ]),
      _: 1
    }, 8, ["show"]));
  }
}), Se = /* @__PURE__ */ Z(me, [["__scopeId", "data-v-d00594a4"]]);
async function te(S) {
  return (await S.json().catch(() => ({ error: S.statusText }))).error ?? `Request failed (${S.status})`;
}
function Ce(S, i = "") {
  const k = i.replace(/\/$/, "");
  return {
    async getContext() {
      const b = await fetch(`${k}/v1/api/member/gatherings/${S}`);
      if (!b.ok) throw new Q(await te(b), b.status);
      return b.json();
    },
    async submitBallot(b) {
      const y = await fetch(`${k}/v1/api/member/gatherings/${S}/ballot`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ballot_content: b })
      });
      if (!y.ok) throw new Q(await te(y), y.status);
      return y.json();
    }
  };
}
export {
  pe as BallotForm,
  Q as HttpError,
  Se as VotingResultsWidget,
  Ve as VotingWidget,
  Ce as createMemberVotingService
};
