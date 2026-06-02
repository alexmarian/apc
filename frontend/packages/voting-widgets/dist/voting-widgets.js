import { defineComponent as K, watch as rt, openBlock as r, createElementBlock as b, Fragment as V, renderList as I, createVNode as o, unref as t, withCtx as e, createBlock as h, createTextVNode as n, toDisplayString as a, createCommentVNode as z, ref as A, computed as U, onMounted as et, createElementVNode as Y, normalizeStyle as dt } from "vue";
import { useI18n as X } from "vue-i18n";
import { NCard as D, NText as N, NRadioGroup as Z, NSpace as M, NRadio as W, NCheckboxGroup as ct, NCheckbox as ft, NButton as J, NSpin as at, NAlert as P, NDescriptions as L, NDescriptionsItem as q, NTag as O, NDivider as lt, NProgress as pt } from "naive-ui";
const gt = { key: 4 }, vt = /* @__PURE__ */ K({
  __name: "BallotForm",
  props: {
    matters: {},
    modelValue: {}
  },
  emits: ["update:modelValue"],
  setup(w, { emit: i }) {
    const { t: x, locale: m } = X({
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
    }), y = w, _ = i;
    rt(
      () => y.matters,
      (d) => {
        const u = { ...y.modelValue };
        let l = !1;
        for (const f of d) {
          const k = String(f.id);
          u[k] === void 0 && (u[k] = f.voting_config.type === "ranking" ? (f.voting_config.options ?? []).map((H) => H.id) : [], l = !0);
        }
        l && _("update:modelValue", u);
      },
      { immediate: !0 }
    );
    function S(d) {
      var u;
      return ((u = y.modelValue[String(d)]) == null ? void 0 : u[0]) ?? "";
    }
    function B(d, u) {
      _("update:modelValue", { ...y.modelValue, [String(d)]: u ? [u] : [] });
    }
    function v(d) {
      return y.modelValue[String(d)] ?? [];
    }
    function C(d, u) {
      _("update:modelValue", { ...y.modelValue, [String(d)]: u.map(String) });
    }
    function T(d) {
      return y.modelValue[String(d)] ?? [];
    }
    function E(d, u, l) {
      const f = [...T(d)], k = u + l;
      k < 0 || k >= f.length || ([f[u], f[k]] = [f[k], f[u]], _("update:modelValue", { ...y.modelValue, [String(d)]: f }));
    }
    function $(d) {
      var l;
      return ((l = m.value) == null ? void 0 : l.slice(0, 2)) === "ru" && d.title_ru ? d.title_ru : d.title;
    }
    function j(d) {
      var l;
      return ((l = m.value) == null ? void 0 : l.slice(0, 2)) === "ru" && d.description_ru ? d.description_ru : d.description;
    }
    function g(d, u) {
      var l, f;
      return ((f = (l = d.voting_config.options) == null ? void 0 : l.find((k) => k.id === u)) == null ? void 0 : f.text) ?? u;
    }
    return (d, u) => (r(), b("div", null, [
      (r(!0), b(V, null, I(w.matters, (l) => (r(), b("div", {
        key: l.id,
        style: { "margin-bottom": "20px" }
      }, [
        o(t(D), { size: "small" }, {
          header: e(() => [
            n(a($(l)), 1)
          ]),
          default: e(() => [
            l.description || l.description_ru ? (r(), h(t(N), {
              key: 0,
              depth: 2,
              style: { display: "block", "margin-bottom": "12px", "font-size": "13px" }
            }, {
              default: e(() => [
                n(a(j(l)), 1)
              ]),
              _: 2
            }, 1024)) : z("", !0),
            l.voting_config.type === "yes_no" ? (r(), h(t(Z), {
              key: 1,
              value: S(l.id),
              "onUpdate:value": (f) => B(l.id, f)
            }, {
              default: e(() => [
                o(t(M), null, {
                  default: e(() => [
                    o(t(W), { value: "yes" }, {
                      default: e(() => [
                        n(a(t(x)("yes")), 1)
                      ]),
                      _: 1
                    }),
                    o(t(W), { value: "no" }, {
                      default: e(() => [
                        n(a(t(x)("no")), 1)
                      ]),
                      _: 1
                    }),
                    l.voting_config.allow_abstention ? (r(), h(t(W), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: e(() => [
                        n(a(t(x)("abstain")), 1)
                      ]),
                      _: 1
                    })) : z("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "single_choice" ? (r(), h(t(Z), {
              key: 2,
              value: S(l.id),
              "onUpdate:value": (f) => B(l.id, f)
            }, {
              default: e(() => [
                o(t(M), { vertical: "" }, {
                  default: e(() => [
                    (r(!0), b(V, null, I(l.voting_config.options ?? [], (f) => (r(), h(t(W), {
                      key: f.id,
                      value: f.id
                    }, {
                      default: e(() => [
                        n(a(f.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128)),
                    l.voting_config.allow_abstention ? (r(), h(t(W), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: e(() => [
                        n(a(t(x)("abstain")), 1)
                      ]),
                      _: 1
                    })) : z("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "multiple_choice" ? (r(), h(t(ct), {
              key: 3,
              value: v(l.id),
              "onUpdate:value": (f) => C(l.id, f)
            }, {
              default: e(() => [
                o(t(M), { vertical: "" }, {
                  default: e(() => [
                    (r(!0), b(V, null, I(l.voting_config.options ?? [], (f) => (r(), h(t(ft), {
                      key: f.id,
                      value: f.id
                    }, {
                      default: e(() => [
                        n(a(f.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128))
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "ranking" ? (r(), b("div", gt, [
              o(t(N), {
                depth: 3,
                style: { "font-size": "12px", "margin-bottom": "10px", display: "block" }
              }, {
                default: e(() => [
                  n(a(t(x)("rankingHint")), 1)
                ]),
                _: 1
              }),
              (r(!0), b(V, null, I(T(l.id), (f, k) => (r(), b("div", {
                key: f,
                style: { display: "flex", "align-items": "center", gap: "8px", "margin-bottom": "6px", padding: "6px 10px", border: "1px solid var(--n-border-color)", "border-radius": "4px" }
              }, [
                o(t(N), {
                  depth: 3,
                  style: { "min-width": "20px", "text-align": "center", "font-weight": "600" }
                }, {
                  default: e(() => [
                    n(a(k + 1), 1)
                  ]),
                  _: 2
                }, 1024),
                o(t(N), { style: { flex: "1" } }, {
                  default: e(() => [
                    n(a(g(l, f)), 1)
                  ]),
                  _: 2
                }, 1024),
                o(t(J), {
                  size: "tiny",
                  disabled: k === 0,
                  onClick: (H) => E(l.id, k, -1)
                }, {
                  default: e(() => [...u[0] || (u[0] = [
                    n("↑", -1)
                  ])]),
                  _: 1
                }, 8, ["disabled", "onClick"]),
                o(t(J), {
                  size: "tiny",
                  disabled: k === T(l.id).length - 1,
                  onClick: (H) => E(l.id, k, 1)
                }, {
                  default: e(() => [...u[1] || (u[1] = [
                    n("↓", -1)
                  ])]),
                  _: 1
                }, 8, ["disabled", "onClick"])
              ]))), 128))
            ])) : z("", !0)
          ]),
          _: 2
        }, 1024)
      ]))), 128))
    ]));
  }
});
class Q extends Error {
  constructor(i, x) {
    super(i), this.status = x, this.name = "HttpError";
  }
}
const bt = { style: { "margin-bottom": "16px" } }, mt = { style: { "margin-top": "8px" } }, yt = { style: { color: "#18a058" } }, _t = /* @__PURE__ */ K({
  __name: "VotingWidget",
  props: {
    service: {},
    initialContext: {}
  },
  setup(w) {
    const { t: i, locale: x } = X({
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
    }), m = w, y = A(!1), _ = A(!1), S = A(null), B = A(null), v = A(null), C = A(null), T = A({}), E = U(
      () => {
        var s;
        return ((s = v.value) == null ? void 0 : s.units.reduce((c, p) => c + p.voting_weight, 0)) ?? 0;
      }
    ), $ = U(() => {
      var c, p;
      const s = (c = v.value) == null ? void 0 : c.gathering;
      if (!s) return 0;
      if (s.voting_mode === "by_weight") {
        const R = s.qualified_units_total_part;
        return R > 0 ? E.value / R * 100 : 0;
      } else {
        const R = s.qualified_units_count;
        return R > 0 ? (((p = v.value) == null ? void 0 : p.units.length) ?? 0) / R * 100 : 0;
      }
    }), j = U(
      () => {
        var s;
        return (((s = v.value) == null ? void 0 : s.units.length) ?? 0) > 0 && v.value.units.every((c) => !c.is_available);
      }
    ), g = U(
      () => {
        var s;
        return (((s = v.value) == null ? void 0 : s.matters) ?? []).filter((c) => !c.is_informative);
      }
    ), d = U(
      () => {
        var s;
        return (((s = v.value) == null ? void 0 : s.matters) ?? []).filter((c) => c.is_informative);
      }
    ), u = U(
      () => g.value.length > 0 && g.value.every((s) => {
        var c;
        return (((c = T.value[String(s.id)]) == null ? void 0 : c.length) ?? 0) > 0;
      })
    ), l = U(() => {
      var s;
      switch ((s = v.value) == null ? void 0 : s.gathering.status) {
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
    function f(s) {
      var p;
      return ((p = x.value) == null ? void 0 : p.slice(0, 2)) === "ru" && s.title_ru ? s.title_ru : s.title;
    }
    function k(s) {
      var p;
      return ((p = x.value) == null ? void 0 : p.slice(0, 2)) === "ru" && s.description_ru ? s.description_ru : s.description;
    }
    function H() {
      const s = {};
      for (const c of g.value) {
        const p = String(c.id);
        s[p] = { matter_id: c.id, values: T.value[p] ?? [] };
      }
      return s;
    }
    function nt(s) {
      if (!C.value) return "—";
      const c = C.value.ballot_content[String(s.id)];
      return !c || c.values.length === 0 ? "—" : s.voting_config.type === "ranking" ? c.values.map((p, R) => {
        var G;
        const F = (G = s.voting_config.options) == null ? void 0 : G.find((st) => st.id === p);
        return `${R + 1}. ${F ? F.text : p}`;
      }).join(", ") : c.values.map((p) => {
        var F;
        if (p === "abstain") return i("abstain");
        if (s.voting_config.type === "yes_no") return i(p === "yes" ? "yes" : "no");
        const R = (F = s.voting_config.options) == null ? void 0 : F.find((G) => G.id === p);
        return R ? R.text : p;
      }).join(", ");
    }
    async function ot() {
      if (m.initialContext) {
        v.value = m.initialContext, m.initialContext.ballot && (C.value = {
          ballot_id: m.initialContext.ballot.ballot_id,
          ballot_hash: m.initialContext.ballot.ballot_hash,
          submitted_at: m.initialContext.ballot.submitted_at,
          ballot_content: m.initialContext.ballot.ballot_content
        });
        return;
      }
      y.value = !0, S.value = null;
      try {
        const s = await m.service.getContext();
        v.value = s, s.ballot && (C.value = {
          ballot_id: s.ballot.ballot_id,
          ballot_hash: s.ballot.ballot_hash,
          submitted_at: s.ballot.submitted_at,
          ballot_content: s.ballot.ballot_content
        });
      } catch (s) {
        S.value = s instanceof Error ? s.message : "Network error";
      } finally {
        y.value = !1;
      }
    }
    async function ut() {
      if (!u.value) return;
      _.value = !0, B.value = null;
      const s = H();
      try {
        const c = await m.service.submitBallot(s);
        C.value = {
          ballot_id: c.ballot_id,
          ballot_hash: c.ballot_hash,
          submitted_at: c.submitted_at,
          ballot_content: c.ballot_content ?? s
        };
      } catch (c) {
        c instanceof Q ? c.status === 409 ? B.value = i("errAlreadySubmitted") : c.status === 400 ? B.value = c.message || i("errInvalidBallot") : B.value = c.message : B.value = c instanceof Error ? c.message : "Network error";
      } finally {
        _.value = !1;
      }
    }
    return et(ot), (s, c) => (r(), h(t(at), { show: y.value }, {
      default: e(() => [
        S.value ? (r(), h(t(P), {
          key: 0,
          type: "error",
          style: { "margin-bottom": "16px" }
        }, {
          default: e(() => [
            n(a(S.value), 1)
          ]),
          _: 1
        })) : z("", !0),
        v.value ? (r(), b(V, { key: 1 }, [
          Y("div", bt, [
            o(t(N), {
              tag: "h2",
              style: { "font-size": "18px", "font-weight": "600", margin: "0 0 4px" }
            }, {
              default: e(() => [
                n(a(v.value.gathering.title), 1)
              ]),
              _: 1
            }),
            v.value.gathering.description ? (r(), h(t(N), {
              key: 0,
              depth: 2,
              style: { "font-size": "14px", display: "block" }
            }, {
              default: e(() => [
                n(a(v.value.gathering.description), 1)
              ]),
              _: 1
            })) : z("", !0)
          ]),
          o(t(D), { style: { "margin-bottom": "16px" } }, {
            default: e(() => [
              o(t(L), {
                column: 3,
                "label-placement": "top",
                size: "small"
              }, {
                default: e(() => [
                  o(t(q), {
                    label: t(i)("owner")
                  }, {
                    default: e(() => [
                      n(a(v.value.owner.name), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(t(q), {
                    label: t(i)("units")
                  }, {
                    default: e(() => [
                      n(a(v.value.units.length), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(t(q), {
                    label: v.value.gathering.voting_mode === "by_weight" ? t(i)("votingWeight") : t(i)("votingUnits")
                  }, {
                    default: e(() => [
                      n(a($.value.toFixed(2)) + "% ", 1)
                    ]),
                    _: 1
                  }, 8, ["label"])
                ]),
                _: 1
              }),
              Y("div", mt, [
                o(t(O), {
                  type: l.value,
                  size: "small"
                }, {
                  default: e(() => [
                    n(a(v.value.gathering.status.toUpperCase()), 1)
                  ]),
                  _: 1
                }, 8, ["type"])
              ])
            ]),
            _: 1
          }),
          v.value.gathering.status !== "active" ? (r(), h(t(P), {
            key: 0,
            type: v.value.gathering.status === "tallied" ? "info" : "warning"
          }, {
            default: e(() => [
              ["draft", "scheduled"].includes(v.value.gathering.status) ? (r(), b(V, { key: 0 }, [
                n(a(t(i)("statusNotStarted")), 1)
              ], 64)) : v.value.gathering.status === "tallied" ? (r(), b(V, { key: 1 }, [
                n(a(t(i)("statusTallied")), 1)
              ], 64)) : (r(), b(V, { key: 2 }, [
                n(a(t(i)("statusClosed")), 1)
              ], 64))
            ]),
            _: 1
          }, 8, ["type"])) : (r(), b(V, { key: 1 }, [
            j.value && !C.value ? (r(), h(t(P), {
              key: 0,
              type: "warning",
              style: { "margin-bottom": "16px" }
            }, {
              default: e(() => [
                n(a(t(i)("unitsAlreadyVoted")), 1)
              ]),
              _: 1
            })) : C.value ? (r(), h(t(D), { key: 1 }, {
              header: e(() => [
                Y("span", yt, "✓ " + a(t(i)("ballotSubmitted")), 1)
              ]),
              default: e(() => [
                o(t(L), {
                  column: 1,
                  "label-placement": "left",
                  size: "small",
                  style: { "margin-bottom": "16px" }
                }, {
                  default: e(() => [
                    o(t(q), {
                      label: t(i)("ballotId")
                    }, {
                      default: e(() => [
                        n(a(C.value.ballot_id), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(t(q), {
                      label: t(i)("verificationHash")
                    }, {
                      default: e(() => [
                        o(t(N), {
                          code: "",
                          style: { "font-size": "11px", "word-break": "break-all" }
                        }, {
                          default: e(() => [
                            n(a(C.value.ballot_hash), 1)
                          ]),
                          _: 1
                        })
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(t(q), {
                      label: t(i)("submittedAt")
                    }, {
                      default: e(() => [
                        n(a(C.value.submitted_at ? new Date(C.value.submitted_at).toLocaleString() : "—"), 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                }),
                o(t(lt), { "title-placement": "left" }, {
                  default: e(() => [
                    n(a(t(i)("yourVotes")), 1)
                  ]),
                  _: 1
                }),
                (r(!0), b(V, null, I(v.value.matters.filter((p) => !p.is_informative), (p) => (r(), b("div", {
                  key: p.id,
                  style: { "margin-bottom": "12px", "padding-left": "4px" }
                }, [
                  o(t(N), {
                    strong: "",
                    style: { display: "block" }
                  }, {
                    default: e(() => [
                      n(a(f(p)), 1)
                    ]),
                    _: 2
                  }, 1024),
                  o(t(N), {
                    depth: 2,
                    style: { "margin-top": "4px", display: "block", "padding-left": "12px" }
                  }, {
                    default: e(() => [
                      n(a(nt(p)), 1)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128))
              ]),
              _: 1
            })) : (r(), h(t(D), { key: 2 }, {
              header: e(() => [
                Y("span", null, a(t(i)("yourBallot")), 1),
                o(t(N), {
                  depth: 3,
                  style: { "font-size": "13px", "margin-left": "8px" }
                }, {
                  default: e(() => [
                    n(" — " + a(v.value.gathering.title), 1)
                  ]),
                  _: 1
                })
              ]),
              default: e(() => [
                B.value ? (r(), h(t(P), {
                  key: 0,
                  type: "error",
                  closable: "",
                  style: { "margin-bottom": "16px" },
                  onClose: c[0] || (c[0] = (p) => B.value = null)
                }, {
                  default: e(() => [
                    n(a(B.value), 1)
                  ]),
                  _: 1
                })) : z("", !0),
                (r(!0), b(V, null, I(d.value, (p) => (r(), b("div", {
                  key: p.id,
                  style: { "margin-bottom": "16px" }
                }, [
                  o(t(D), {
                    size: "small",
                    embedded: ""
                  }, {
                    header: e(() => [
                      o(t(N), { style: { "font-size": "14px" } }, {
                        default: e(() => [
                          n(a(f(p)), 1)
                        ]),
                        _: 2
                      }, 1024),
                      o(t(O), {
                        size: "tiny",
                        style: { "margin-left": "8px" }
                      }, {
                        default: e(() => [
                          n(a(t(i)("informational")), 1)
                        ]),
                        _: 1
                      })
                    ]),
                    default: e(() => [
                      o(t(N), {
                        depth: 2,
                        style: { "font-size": "13px" }
                      }, {
                        default: e(() => [
                          n(a(k(p)), 1)
                        ]),
                        _: 2
                      }, 1024)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128)),
                o(vt, {
                  matters: g.value,
                  modelValue: T.value,
                  "onUpdate:modelValue": c[1] || (c[1] = (p) => T.value = p)
                }, null, 8, ["matters", "modelValue"]),
                o(t(M), {
                  justify: "end",
                  style: { "margin-top": "8px" }
                }, {
                  default: e(() => [
                    o(t(J), {
                      type: "primary",
                      loading: _.value,
                      disabled: !u.value,
                      onClick: ut
                    }, {
                      default: e(() => [
                        n(a(t(i)("submitBallot")), 1)
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
        ], 64)) : z("", !0)
      ]),
      _: 1
    }, 8, ["show"]));
  }
}), it = (w, i) => {
  const x = w.__vccOpts || w;
  for (const [m, y] of i)
    x[m] = y;
  return x;
}, wt = /* @__PURE__ */ it(_t, [["__scopeId", "data-v-a8ef0700"]]), ht = /* @__PURE__ */ K({
  __name: "VotingResultsWidget",
  props: {
    service: {},
    initialContext: {}
  },
  setup(w) {
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
    }), x = w, m = A(!1), y = A(null), _ = A(null), S = A(null), B = U(() => {
      var g;
      switch ((g = _.value) == null ? void 0 : g.gathering.status) {
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
    function v(g) {
      var u;
      if (!((u = _.value) != null && u.ballot)) return !1;
      const d = _.value.ballot.ballot_content[String(g)];
      return !!d && d.values.length > 0;
    }
    function C(g, d) {
      var l;
      if (!((l = _.value) != null && l.ballot)) return !1;
      const u = _.value.ballot.ballot_content[String(g)];
      return !!u && u.values.includes(d);
    }
    function T(g, d) {
      var l;
      if (g === "abstain") return i("abstain");
      if (d.type === "yes_no") return i(g === "yes" ? "yes" : "no");
      const u = (l = d.options) == null ? void 0 : l.find((f) => f.id === g);
      return u ? u.text : g;
    }
    function E(g) {
      return [...g.votes].sort((d, u) => u.vote_count - d.vote_count);
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
      if (x.initialContext) {
        _.value = x.initialContext, S.value = x.initialContext.results ?? null;
        return;
      }
      m.value = !0, y.value = null;
      try {
        const g = await x.service.getContext();
        _.value = g, S.value = g.results ?? null;
      } catch (g) {
        y.value = g instanceof Error ? g.message : "Network error";
      } finally {
        m.value = !1;
      }
    }
    return et(j), (g, d) => (r(), h(t(at), { show: m.value }, {
      default: e(() => [
        y.value ? (r(), h(t(P), {
          key: 0,
          type: "error",
          style: { "margin-bottom": "16px" }
        }, {
          default: e(() => [
            n(a(y.value), 1)
          ]),
          _: 1
        })) : z("", !0),
        _.value ? (r(), b(V, { key: 1 }, [
          o(t(D), { style: { "margin-bottom": "16px" } }, {
            default: e(() => [
              o(t(L), {
                column: 2,
                "label-placement": "top",
                size: "small"
              }, {
                default: e(() => [
                  o(t(q), {
                    label: t(i)("gathering")
                  }, {
                    default: e(() => [
                      n(a(_.value.gathering.title), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  o(t(q), {
                    label: t(i)("status")
                  }, {
                    default: e(() => [
                      o(t(O), {
                        type: B.value,
                        size: "small"
                      }, {
                        default: e(() => [
                          n(a(_.value.gathering.status.toUpperCase()), 1)
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
          S.value ? (r(), b(V, { key: 1 }, [
            o(t(D), {
              size: "small",
              style: { "margin-bottom": "16px" },
              title: t(i)("participationSummary")
            }, {
              default: e(() => [
                o(t(L), {
                  column: 3,
                  "label-placement": "top",
                  size: "small"
                }, {
                  default: e(() => [
                    o(t(q), {
                      label: t(i)("participated")
                    }, {
                      default: e(() => [
                        n(a(S.value.statistics.participating_units) + " " + a(t(i)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(t(q), {
                      label: t(i)("voted")
                    }, {
                      default: e(() => [
                        n(a(S.value.statistics.voted_units) + " " + a(t(i)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    o(t(q), {
                      label: t(i)("participationRate")
                    }, {
                      default: e(() => [
                        n(a(S.value.statistics.participation_rate.toFixed(1)) + "% ", 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["title"]),
            (r(!0), b(V, null, I(S.value.results, (u) => (r(), b("div", {
              key: u.matter_id,
              style: { "margin-bottom": "16px" }
            }, [
              o(t(D), { size: "small" }, {
                header: e(() => [
                  o(t(M), {
                    align: "center",
                    justify: "space-between",
                    style: { "flex-wrap": "wrap", gap: "4px" }
                  }, {
                    default: e(() => [
                      o(t(N), { strong: "" }, {
                        default: e(() => [
                          n(a(u.matter_title), 1)
                        ]),
                        _: 2
                      }, 1024),
                      o(t(O), {
                        type: u.is_passed ? "success" : "error",
                        size: "small"
                      }, {
                        default: e(() => [
                          n(a(u.is_passed ? t(i)("passed") : t(i)("failed")), 1)
                        ]),
                        _: 2
                      }, 1032, ["type"])
                    ]),
                    _: 2
                  }, 1024)
                ]),
                default: e(() => [
                  v(u.matter_id) ? (r(), h(t(P), {
                    key: 0,
                    type: "success",
                    size: "small",
                    style: { "margin-bottom": "12px" }
                  }, {
                    default: e(() => [
                      n(a(t(i)("yourVoteCounted")), 1)
                    ]),
                    _: 1
                  })) : _.value.ballot ? (r(), h(t(P), {
                    key: 1,
                    type: "default",
                    size: "small",
                    style: { "margin-bottom": "12px" }
                  }, {
                    default: e(() => [
                      n(a(t(i)("didNotVote")), 1)
                    ]),
                    _: 1
                  })) : z("", !0),
                  (r(!0), b(V, null, I(E(u), (l) => (r(), b("div", {
                    key: l.choice,
                    style: { "margin-bottom": "10px" }
                  }, [
                    o(t(M), {
                      align: "center",
                      justify: "space-between",
                      style: { "margin-bottom": "4px" }
                    }, {
                      default: e(() => [
                        o(t(M), {
                          align: "center",
                          size: "small"
                        }, {
                          default: e(() => [
                            o(t(N), {
                              style: dt(C(u.matter_id, l.choice) ? "font-weight:600;color:#18a058" : "")
                            }, {
                              default: e(() => [
                                n(a(T(l.choice, u.voting_config)), 1)
                              ]),
                              _: 2
                            }, 1032, ["style"]),
                            C(u.matter_id, l.choice) ? (r(), h(t(O), {
                              key: 0,
                              type: "success",
                              size: "tiny"
                            }, {
                              default: e(() => [
                                n(a(t(i)("yourVote")), 1)
                              ]),
                              _: 1
                            })) : z("", !0)
                          ]),
                          _: 2
                        }, 1024),
                        o(t(N), {
                          depth: 2,
                          style: { "font-size": "12px" }
                        }, {
                          default: e(() => {
                            var f, k;
                            return [
                              n(a(l.vote_count) + " " + a(l.vote_count !== 1 ? t(i)("votes") : t(i)("vote")) + " (" + a(l.percentage.toFixed(1)) + "%", 1),
                              ((k = (f = S.value) == null ? void 0 : f.statistics) == null ? void 0 : k.voting_mode) === "by_weight" ? (r(), b(V, { key: 0 }, [
                                n(" · " + a(t(i)("weight")) + ": " + a(l.weight_percentage.toFixed(1)) + "%", 1)
                              ], 64)) : z("", !0),
                              d[1] || (d[1] = n(") ", -1))
                            ];
                          }),
                          _: 2
                        }, 1024)
                      ]),
                      _: 2
                    }, 1024),
                    o(t(pt), {
                      type: "line",
                      percentage: l.percentage,
                      status: $(l.choice, u),
                      "show-indicator": !1,
                      height: 8,
                      "border-radius": 4
                    }, null, 8, ["percentage", "status"])
                  ]))), 128)),
                  o(t(lt), { style: { margin: "8px 0" } }),
                  o(t(N), {
                    depth: 3,
                    style: { "font-size": "12px" }
                  }, {
                    default: e(() => {
                      var l;
                      return [
                        n(a(t(i)("quorum")) + ": " + a((l = u.quorum_info) != null && l.met ? t(i)("quorumMet") : t(i)("quorumNotMet")) + " ", 1),
                        u.quorum_info ? (r(), b(V, { key: 0 }, [
                          n(" — " + a(u.quorum_info.achieved_percentage.toFixed(1)) + "% " + a(t(i)("of")) + " " + a(u.quorum_info.required_percentage) + "% " + a(t(i)("required")), 1)
                        ], 64)) : z("", !0)
                      ];
                    }),
                    _: 2
                  }, 1024)
                ]),
                _: 2
              }, 1024)
            ]))), 128))
          ], 64)) : (r(), h(t(P), {
            key: 0,
            type: "info"
          }, {
            default: e(() => [
              n(a(t(i)("notAvailable")) + " ", 1),
              Y("strong", null, a(_.value.gathering.status), 1),
              d[0] || (d[0] = n(". ", -1)),
              _.value.gathering.status !== "tallied" ? (r(), b(V, { key: 0 }, [
                n(a(t(i)("willBePublished")), 1)
              ], 64)) : z("", !0)
            ]),
            _: 1
          }))
        ], 64)) : z("", !0)
      ]),
      _: 1
    }, 8, ["show"]));
  }
}), St = /* @__PURE__ */ it(ht, [["__scopeId", "data-v-f9ffa84d"]]);
async function tt(w) {
  return (await w.json().catch(() => ({ error: w.statusText }))).error ?? `Request failed (${w.status})`;
}
function Ct(w, i = "") {
  const x = i.replace(/\/$/, "");
  return {
    async getContext() {
      const m = await fetch(`${x}/v1/api/member/gatherings/${w}`);
      if (!m.ok) throw new Q(await tt(m), m.status);
      return m.json();
    },
    async submitBallot(m) {
      const y = await fetch(`${x}/v1/api/member/gatherings/${w}/ballot`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ballot_content: m })
      });
      if (!y.ok) throw new Q(await tt(y), y.status);
      return y.json();
    }
  };
}
export {
  vt as BallotForm,
  Q as HttpError,
  St as VotingResultsWidget,
  wt as VotingWidget,
  Ct as createMemberVotingService
};
