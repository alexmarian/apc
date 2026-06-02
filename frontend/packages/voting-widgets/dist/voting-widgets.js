import { defineComponent as J, watch as it, openBlock as u, createElementBlock as g, Fragment as w, renderList as T, createVNode as i, unref as t, withCtx as e, createBlock as h, createTextVNode as n, toDisplayString as a, createCommentVNode as B, ref as q, reactive as ot, computed as D, onMounted as Z, createElementVNode as F, isRef as nt, normalizeStyle as st } from "vue";
import { useI18n as Q } from "vue-i18n";
import { NCard as I, NText as N, NRadioGroup as K, NSpace as P, NRadio as U, NCheckboxGroup as ut, NCheckbox as rt, NButton as G, NSpin as tt, NAlert as M, NDescriptions as Y, NDescriptionsItem as A, NTag as j, NDivider as et, NProgress as dt } from "naive-ui";
const ct = { key: 4 }, ft = /* @__PURE__ */ J({
  __name: "BallotForm",
  props: {
    matters: {},
    modelValue: {}
  },
  emits: ["update:modelValue"],
  setup(S, { emit: o }) {
    const { t: k } = Q({
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
    }), v = S, y = o;
    it(
      () => v.matters,
      (m) => {
        const f = { ...v.modelValue };
        let l = !1;
        for (const s of m) {
          const r = String(s.id);
          f[r] === void 0 && (f[r] = s.voting_config.type === "ranking" ? (s.voting_config.options ?? []).map((b) => b.id) : [], l = !0);
        }
        l && y("update:modelValue", f);
      },
      { immediate: !0 }
    );
    function _(m) {
      var f;
      return ((f = v.modelValue[String(m)]) == null ? void 0 : f[0]) ?? "";
    }
    function V(m, f) {
      y("update:modelValue", { ...v.modelValue, [String(m)]: f ? [f] : [] });
    }
    function x(m) {
      return v.modelValue[String(m)] ?? [];
    }
    function C(m, f) {
      y("update:modelValue", { ...v.modelValue, [String(m)]: f.map(String) });
    }
    function z(m) {
      return v.modelValue[String(m)] ?? [];
    }
    function E(m, f, l) {
      const s = [...z(m)], r = f + l;
      r < 0 || r >= s.length || ([s[f], s[r]] = [s[r], s[f]], y("update:modelValue", { ...v.modelValue, [String(m)]: s }));
    }
    function R(m, f) {
      var l, s;
      return ((s = (l = m.voting_config.options) == null ? void 0 : l.find((r) => r.id === f)) == null ? void 0 : s.text) ?? f;
    }
    return (m, f) => (u(), g("div", null, [
      (u(!0), g(w, null, T(S.matters, (l) => (u(), g("div", {
        key: l.id,
        style: { "margin-bottom": "20px" }
      }, [
        i(t(I), { size: "small" }, {
          header: e(() => [
            n(a(l.title), 1)
          ]),
          default: e(() => [
            l.description ? (u(), h(t(N), {
              key: 0,
              depth: 2,
              style: { display: "block", "margin-bottom": "12px", "font-size": "13px" }
            }, {
              default: e(() => [
                n(a(l.description), 1)
              ]),
              _: 2
            }, 1024)) : B("", !0),
            l.voting_config.type === "yes_no" ? (u(), h(t(K), {
              key: 1,
              value: _(l.id),
              "onUpdate:value": (s) => V(l.id, s)
            }, {
              default: e(() => [
                i(t(P), null, {
                  default: e(() => [
                    i(t(U), { value: "yes" }, {
                      default: e(() => [
                        n(a(t(k)("yes")), 1)
                      ]),
                      _: 1
                    }),
                    i(t(U), { value: "no" }, {
                      default: e(() => [
                        n(a(t(k)("no")), 1)
                      ]),
                      _: 1
                    }),
                    l.voting_config.allow_abstention ? (u(), h(t(U), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: e(() => [
                        n(a(t(k)("abstain")), 1)
                      ]),
                      _: 1
                    })) : B("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "single_choice" ? (u(), h(t(K), {
              key: 2,
              value: _(l.id),
              "onUpdate:value": (s) => V(l.id, s)
            }, {
              default: e(() => [
                i(t(P), { vertical: "" }, {
                  default: e(() => [
                    (u(!0), g(w, null, T(l.voting_config.options ?? [], (s) => (u(), h(t(U), {
                      key: s.id,
                      value: s.id
                    }, {
                      default: e(() => [
                        n(a(s.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128)),
                    l.voting_config.allow_abstention ? (u(), h(t(U), {
                      key: 0,
                      value: "abstain"
                    }, {
                      default: e(() => [
                        n(a(t(k)("abstain")), 1)
                      ]),
                      _: 1
                    })) : B("", !0)
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "multiple_choice" ? (u(), h(t(ut), {
              key: 3,
              value: x(l.id),
              "onUpdate:value": (s) => C(l.id, s)
            }, {
              default: e(() => [
                i(t(P), { vertical: "" }, {
                  default: e(() => [
                    (u(!0), g(w, null, T(l.voting_config.options ?? [], (s) => (u(), h(t(rt), {
                      key: s.id,
                      value: s.id
                    }, {
                      default: e(() => [
                        n(a(s.text), 1)
                      ]),
                      _: 2
                    }, 1032, ["value"]))), 128))
                  ]),
                  _: 2
                }, 1024)
              ]),
              _: 2
            }, 1032, ["value", "onUpdate:value"])) : l.voting_config.type === "ranking" ? (u(), g("div", ct, [
              i(t(N), {
                depth: 3,
                style: { "font-size": "12px", "margin-bottom": "10px", display: "block" }
              }, {
                default: e(() => [
                  n(a(t(k)("rankingHint")), 1)
                ]),
                _: 1
              }),
              (u(!0), g(w, null, T(z(l.id), (s, r) => (u(), g("div", {
                key: s,
                style: { display: "flex", "align-items": "center", gap: "8px", "margin-bottom": "6px", padding: "6px 10px", border: "1px solid var(--n-border-color)", "border-radius": "4px" }
              }, [
                i(t(N), {
                  depth: 3,
                  style: { "min-width": "20px", "text-align": "center", "font-weight": "600" }
                }, {
                  default: e(() => [
                    n(a(r + 1), 1)
                  ]),
                  _: 2
                }, 1024),
                i(t(N), { style: { flex: "1" } }, {
                  default: e(() => [
                    n(a(R(l, s)), 1)
                  ]),
                  _: 2
                }, 1024),
                i(t(G), {
                  size: "tiny",
                  disabled: r === 0,
                  onClick: (b) => E(l.id, r, -1)
                }, {
                  default: e(() => [...f[0] || (f[0] = [
                    n("↑", -1)
                  ])]),
                  _: 1
                }, 8, ["disabled", "onClick"]),
                i(t(G), {
                  size: "tiny",
                  disabled: r === z(l.id).length - 1,
                  onClick: (b) => E(l.id, r, 1)
                }, {
                  default: e(() => [...f[1] || (f[1] = [
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
});
class L extends Error {
  constructor(o, k) {
    super(o), this.status = k, this.name = "HttpError";
  }
}
const pt = { style: { "margin-top": "8px" } }, vt = { style: { color: "#18a058" } }, gt = /* @__PURE__ */ J({
  __name: "VotingWidget",
  props: {
    service: {}
  },
  setup(S) {
    const { t: o } = Q({
      useScope: "local",
      messages: {
        en: {
          owner: "Owner",
          units: "Units",
          votingWeight: "Voting weight",
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
          errInvalidBallot: "Invalid ballot."
        },
        ro: {
          owner: "Proprietar",
          units: "Unități",
          votingWeight: "Pondere de vot",
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
          errInvalidBallot: "Buletin de vot invalid."
        },
        ru: {
          owner: "Владелец",
          units: "Единицы",
          votingWeight: "Вес голоса",
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
          errInvalidBallot: "Недействительный бюллетень."
        }
      }
    }), k = S, v = q(!1), y = q(!1), _ = q(null), V = q(null), x = q(null), C = q(null);
    let z = ot({});
    const E = D(
      () => {
        var d;
        return ((d = x.value) == null ? void 0 : d.units.reduce((c, p) => c + p.voting_weight, 0)) ?? 0;
      }
    ), R = D(
      () => {
        var d;
        return (((d = x.value) == null ? void 0 : d.matters) ?? []).filter((c) => !c.is_informative);
      }
    ), m = D(
      () => {
        var d;
        return (((d = x.value) == null ? void 0 : d.matters) ?? []).filter((c) => c.is_informative);
      }
    ), f = D(
      () => R.value.length > 0 && R.value.every((d) => {
        var c;
        return (((c = z[String(d.id)]) == null ? void 0 : c.length) ?? 0) > 0;
      })
    ), l = D(() => {
      var d;
      switch ((d = x.value) == null ? void 0 : d.gathering.status) {
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
    function s() {
      const d = {};
      for (const c of R.value) {
        const p = String(c.id);
        d[p] = { matter_id: c.id, values: z[p] ?? [] };
      }
      return d;
    }
    function r(d) {
      if (!C.value) return "—";
      const c = C.value.ballot_content[String(d.id)];
      return !c || c.values.length === 0 ? "—" : d.voting_config.type === "ranking" ? c.values.map((p, H) => {
        var W;
        const $ = (W = d.voting_config.options) == null ? void 0 : W.find((lt) => lt.id === p);
        return `${H + 1}. ${$ ? $.text : p}`;
      }).join(", ") : c.values.map((p) => {
        var $;
        if (p === "abstain") return o("abstain");
        if (d.voting_config.type === "yes_no") return o(p === "yes" ? "yes" : "no");
        const H = ($ = d.voting_config.options) == null ? void 0 : $.find((W) => W.id === p);
        return H ? H.text : p;
      }).join(", ");
    }
    async function b() {
      v.value = !0, _.value = null;
      try {
        const d = await k.service.getContext();
        x.value = d, d.ballot && (C.value = {
          ballot_id: d.ballot.ballot_id,
          ballot_hash: d.ballot.ballot_hash,
          submitted_at: d.ballot.submitted_at,
          ballot_content: d.ballot.ballot_content
        });
      } catch (d) {
        _.value = d instanceof Error ? d.message : "Network error";
      } finally {
        v.value = !1;
      }
    }
    async function O() {
      if (!f.value) return;
      y.value = !0, V.value = null;
      const d = s();
      try {
        const c = await k.service.submitBallot(d);
        C.value = {
          ballot_id: c.ballot_id,
          ballot_hash: c.ballot_hash,
          submitted_at: c.submitted_at,
          ballot_content: c.ballot_content ?? d
        };
      } catch (c) {
        c instanceof L ? c.status === 409 ? V.value = o("errAlreadySubmitted") : c.status === 400 ? V.value = c.message || o("errInvalidBallot") : V.value = c.message : V.value = c instanceof Error ? c.message : "Network error";
      } finally {
        y.value = !1;
      }
    }
    return Z(b), (d, c) => (u(), h(t(tt), { show: v.value }, {
      default: e(() => [
        _.value ? (u(), h(t(M), {
          key: 0,
          type: "error",
          style: { "margin-bottom": "16px" }
        }, {
          default: e(() => [
            n(a(_.value), 1)
          ]),
          _: 1
        })) : B("", !0),
        x.value ? (u(), g(w, { key: 1 }, [
          i(t(I), { style: { "margin-bottom": "16px" } }, {
            default: e(() => [
              i(t(Y), {
                column: 3,
                "label-placement": "top",
                size: "small"
              }, {
                default: e(() => [
                  i(t(A), {
                    label: t(o)("owner")
                  }, {
                    default: e(() => [
                      n(a(x.value.owner.name), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  i(t(A), {
                    label: t(o)("units")
                  }, {
                    default: e(() => [
                      n(a(x.value.units.length), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  i(t(A), {
                    label: t(o)("votingWeight")
                  }, {
                    default: e(() => [
                      n(a(E.value.toFixed(4)), 1)
                    ]),
                    _: 1
                  }, 8, ["label"])
                ]),
                _: 1
              }),
              F("div", pt, [
                i(t(j), {
                  type: l.value,
                  size: "small"
                }, {
                  default: e(() => [
                    n(a(x.value.gathering.status.toUpperCase()), 1)
                  ]),
                  _: 1
                }, 8, ["type"])
              ])
            ]),
            _: 1
          }),
          x.value.gathering.status !== "active" ? (u(), h(t(M), {
            key: 0,
            type: x.value.gathering.status === "tallied" ? "info" : "warning"
          }, {
            default: e(() => [
              ["draft", "scheduled"].includes(x.value.gathering.status) ? (u(), g(w, { key: 0 }, [
                n(a(t(o)("statusNotStarted")), 1)
              ], 64)) : x.value.gathering.status === "tallied" ? (u(), g(w, { key: 1 }, [
                n(a(t(o)("statusTallied")), 1)
              ], 64)) : (u(), g(w, { key: 2 }, [
                n(a(t(o)("statusClosed")), 1)
              ], 64))
            ]),
            _: 1
          }, 8, ["type"])) : (u(), g(w, { key: 1 }, [
            C.value ? (u(), h(t(I), { key: 0 }, {
              header: e(() => [
                F("span", vt, "✓ " + a(t(o)("ballotSubmitted")), 1)
              ]),
              default: e(() => [
                i(t(Y), {
                  column: 1,
                  "label-placement": "left",
                  size: "small",
                  style: { "margin-bottom": "16px" }
                }, {
                  default: e(() => [
                    i(t(A), {
                      label: t(o)("ballotId")
                    }, {
                      default: e(() => [
                        n(a(C.value.ballot_id), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    i(t(A), {
                      label: t(o)("verificationHash")
                    }, {
                      default: e(() => [
                        i(t(N), {
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
                    i(t(A), {
                      label: t(o)("submittedAt")
                    }, {
                      default: e(() => [
                        n(a(C.value.submitted_at ? new Date(C.value.submitted_at).toLocaleString() : "—"), 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                }),
                i(t(et), { "title-placement": "left" }, {
                  default: e(() => [
                    n(a(t(o)("yourVotes")), 1)
                  ]),
                  _: 1
                }),
                (u(!0), g(w, null, T(x.value.matters.filter((p) => !p.is_informative), (p) => (u(), g("div", {
                  key: p.id,
                  style: { "margin-bottom": "12px", "padding-left": "4px" }
                }, [
                  i(t(N), {
                    strong: "",
                    style: { display: "block" }
                  }, {
                    default: e(() => [
                      n(a(p.title), 1)
                    ]),
                    _: 2
                  }, 1024),
                  i(t(N), {
                    depth: 2,
                    style: { "margin-top": "4px", display: "block", "padding-left": "12px" }
                  }, {
                    default: e(() => [
                      n(a(r(p)), 1)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128))
              ]),
              _: 1
            })) : (u(), h(t(I), { key: 1 }, {
              header: e(() => [
                F("span", null, a(t(o)("yourBallot")), 1),
                i(t(N), {
                  depth: 3,
                  style: { "font-size": "13px", "margin-left": "8px" }
                }, {
                  default: e(() => [
                    n(" — " + a(x.value.gathering.title), 1)
                  ]),
                  _: 1
                })
              ]),
              default: e(() => [
                V.value ? (u(), h(t(M), {
                  key: 0,
                  type: "error",
                  closable: "",
                  style: { "margin-bottom": "16px" },
                  onClose: c[0] || (c[0] = (p) => V.value = null)
                }, {
                  default: e(() => [
                    n(a(V.value), 1)
                  ]),
                  _: 1
                })) : B("", !0),
                (u(!0), g(w, null, T(m.value, (p) => (u(), g("div", {
                  key: p.id,
                  style: { "margin-bottom": "16px" }
                }, [
                  i(t(I), {
                    size: "small",
                    embedded: ""
                  }, {
                    header: e(() => [
                      i(t(N), { style: { "font-size": "14px" } }, {
                        default: e(() => [
                          n(a(p.title), 1)
                        ]),
                        _: 2
                      }, 1024),
                      i(t(j), {
                        size: "tiny",
                        style: { "margin-left": "8px" }
                      }, {
                        default: e(() => [
                          n(a(t(o)("informational")), 1)
                        ]),
                        _: 1
                      })
                    ]),
                    default: e(() => [
                      i(t(N), {
                        depth: 2,
                        style: { "font-size": "13px" }
                      }, {
                        default: e(() => [
                          n(a(p.description), 1)
                        ]),
                        _: 2
                      }, 1024)
                    ]),
                    _: 2
                  }, 1024)
                ]))), 128)),
                i(ft, {
                  matters: R.value,
                  modelValue: t(z),
                  "onUpdate:modelValue": c[1] || (c[1] = (p) => nt(z) ? z.value = p : z = p)
                }, null, 8, ["matters", "modelValue"]),
                i(t(P), {
                  justify: "end",
                  style: { "margin-top": "8px" }
                }, {
                  default: e(() => [
                    i(t(G), {
                      type: "primary",
                      loading: y.value,
                      disabled: !f.value,
                      onClick: O
                    }, {
                      default: e(() => [
                        n(a(t(o)("submitBallot")), 1)
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
}), at = (S, o) => {
  const k = S.__vccOpts || S;
  for (const [v, y] of o)
    k[v] = y;
  return k;
}, ht = /* @__PURE__ */ at(gt, [["__scopeId", "data-v-2c9825bd"]]), bt = /* @__PURE__ */ J({
  __name: "VotingResultsWidget",
  props: {
    service: {}
  },
  setup(S) {
    const { t: o } = Q({
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
          abstain: "Abstain",
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
          abstain: "Abținere",
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
          abstain: "Воздержаться",
          quorum: "Кворум",
          quorumMet: "Достигнут",
          quorumNotMet: "Не достигнут",
          of: "из",
          required: "требуется",
          notAvailable: "Результаты пока недоступны. Текущий статус:",
          willBePublished: "Результаты будут опубликованы после подсчёта голосов."
        }
      }
    }), k = S, v = q(!1), y = q(null), _ = q(null), V = q(null), x = D(() => {
      var l;
      switch ((l = _.value) == null ? void 0 : l.gathering.status) {
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
    function C(l) {
      var r;
      if (!((r = _.value) != null && r.ballot)) return !1;
      const s = _.value.ballot.ballot_content[String(l)];
      return !!s && s.values.length > 0;
    }
    function z(l, s) {
      var b;
      if (!((b = _.value) != null && b.ballot)) return !1;
      const r = _.value.ballot.ballot_content[String(l)];
      return !!r && r.values.includes(s);
    }
    function E(l, s) {
      var b;
      if (l === "abstain") return o("abstain");
      if (s.type === "yes_no") return l === "yes" ? "Yes" : "No";
      const r = (b = s.options) == null ? void 0 : b.find((O) => O.id === l);
      return r ? r.text : l;
    }
    function R(l) {
      return [...l.votes].sort((s, r) => r.vote_count - s.vote_count);
    }
    function m(l, s) {
      if (l === "abstain") return "warning";
      if (s.voting_config.type === "yes_no") {
        if (l === "yes") return s.is_passed ? "success" : "default";
        if (l === "no") return s.is_passed ? "default" : "error";
      }
      return "default";
    }
    async function f() {
      v.value = !0, y.value = null;
      try {
        const l = await k.service.getContext();
        _.value = l, V.value = l.results ?? null;
      } catch (l) {
        y.value = l instanceof Error ? l.message : "Network error";
      } finally {
        v.value = !1;
      }
    }
    return Z(f), (l, s) => (u(), h(t(tt), { show: v.value }, {
      default: e(() => [
        y.value ? (u(), h(t(M), {
          key: 0,
          type: "error",
          style: { "margin-bottom": "16px" }
        }, {
          default: e(() => [
            n(a(y.value), 1)
          ]),
          _: 1
        })) : B("", !0),
        _.value ? (u(), g(w, { key: 1 }, [
          i(t(I), { style: { "margin-bottom": "16px" } }, {
            default: e(() => [
              i(t(Y), {
                column: 2,
                "label-placement": "top",
                size: "small"
              }, {
                default: e(() => [
                  i(t(A), {
                    label: t(o)("gathering")
                  }, {
                    default: e(() => [
                      n(a(_.value.gathering.title), 1)
                    ]),
                    _: 1
                  }, 8, ["label"]),
                  i(t(A), {
                    label: t(o)("status")
                  }, {
                    default: e(() => [
                      i(t(j), {
                        type: x.value,
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
          V.value ? (u(), g(w, { key: 1 }, [
            i(t(I), {
              size: "small",
              style: { "margin-bottom": "16px" },
              title: t(o)("participationSummary")
            }, {
              default: e(() => [
                i(t(Y), {
                  column: 3,
                  "label-placement": "top",
                  size: "small"
                }, {
                  default: e(() => [
                    i(t(A), {
                      label: t(o)("participated")
                    }, {
                      default: e(() => [
                        n(a(V.value.statistics.participating_units) + " " + a(t(o)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    i(t(A), {
                      label: t(o)("voted")
                    }, {
                      default: e(() => [
                        n(a(V.value.statistics.voted_units) + " " + a(t(o)("units")), 1)
                      ]),
                      _: 1
                    }, 8, ["label"]),
                    i(t(A), {
                      label: t(o)("participationRate")
                    }, {
                      default: e(() => [
                        n(a(V.value.statistics.participation_rate.toFixed(1)) + "% ", 1)
                      ]),
                      _: 1
                    }, 8, ["label"])
                  ]),
                  _: 1
                })
              ]),
              _: 1
            }, 8, ["title"]),
            (u(!0), g(w, null, T(V.value.results, (r) => (u(), g("div", {
              key: r.matter_id,
              style: { "margin-bottom": "16px" }
            }, [
              i(t(I), { size: "small" }, {
                header: e(() => [
                  i(t(P), {
                    align: "center",
                    justify: "space-between",
                    style: { "flex-wrap": "wrap", gap: "4px" }
                  }, {
                    default: e(() => [
                      i(t(N), { strong: "" }, {
                        default: e(() => [
                          n(a(r.matter_title), 1)
                        ]),
                        _: 2
                      }, 1024),
                      i(t(j), {
                        type: r.is_passed ? "success" : "error",
                        size: "small"
                      }, {
                        default: e(() => [
                          n(a(r.is_passed ? t(o)("passed") : t(o)("failed")), 1)
                        ]),
                        _: 2
                      }, 1032, ["type"])
                    ]),
                    _: 2
                  }, 1024)
                ]),
                default: e(() => [
                  C(r.matter_id) ? (u(), h(t(M), {
                    key: 0,
                    type: "success",
                    size: "small",
                    style: { "margin-bottom": "12px" }
                  }, {
                    default: e(() => [
                      n(a(t(o)("yourVoteCounted")), 1)
                    ]),
                    _: 1
                  })) : _.value.ballot ? (u(), h(t(M), {
                    key: 1,
                    type: "default",
                    size: "small",
                    style: { "margin-bottom": "12px" }
                  }, {
                    default: e(() => [
                      n(a(t(o)("didNotVote")), 1)
                    ]),
                    _: 1
                  })) : B("", !0),
                  (u(!0), g(w, null, T(R(r), (b) => (u(), g("div", {
                    key: b.choice,
                    style: { "margin-bottom": "10px" }
                  }, [
                    i(t(P), {
                      align: "center",
                      justify: "space-between",
                      style: { "margin-bottom": "4px" }
                    }, {
                      default: e(() => [
                        i(t(P), {
                          align: "center",
                          size: "small"
                        }, {
                          default: e(() => [
                            i(t(N), {
                              style: st(z(r.matter_id, b.choice) ? "font-weight:600;color:#18a058" : "")
                            }, {
                              default: e(() => [
                                n(a(E(b.choice, r.voting_config)), 1)
                              ]),
                              _: 2
                            }, 1032, ["style"]),
                            z(r.matter_id, b.choice) ? (u(), h(t(j), {
                              key: 0,
                              type: "success",
                              size: "tiny"
                            }, {
                              default: e(() => [
                                n(a(t(o)("yourVote")), 1)
                              ]),
                              _: 1
                            })) : B("", !0)
                          ]),
                          _: 2
                        }, 1024),
                        i(t(N), {
                          depth: 2,
                          style: { "font-size": "12px" }
                        }, {
                          default: e(() => [
                            n(a(b.vote_count) + " " + a(b.vote_count !== 1 ? t(o)("votes") : t(o)("vote")) + " (" + a(b.percentage.toFixed(1)) + "%) ", 1)
                          ]),
                          _: 2
                        }, 1024)
                      ]),
                      _: 2
                    }, 1024),
                    i(t(dt), {
                      type: "line",
                      percentage: b.percentage,
                      status: m(b.choice, r),
                      "show-indicator": !1,
                      height: 8,
                      "border-radius": 4
                    }, null, 8, ["percentage", "status"])
                  ]))), 128)),
                  i(t(et), { style: { margin: "8px 0" } }),
                  i(t(N), {
                    depth: 3,
                    style: { "font-size": "12px" }
                  }, {
                    default: e(() => {
                      var b;
                      return [
                        n(a(t(o)("quorum")) + ": " + a((b = r.quorum_info) != null && b.met ? t(o)("quorumMet") : t(o)("quorumNotMet")) + " ", 1),
                        r.quorum_info ? (u(), g(w, { key: 0 }, [
                          n(" — " + a(r.quorum_info.achieved_percentage.toFixed(1)) + "% " + a(t(o)("of")) + " " + a(r.quorum_info.required_percentage) + "% " + a(t(o)("required")), 1)
                        ], 64)) : B("", !0)
                      ];
                    }),
                    _: 2
                  }, 1024)
                ]),
                _: 2
              }, 1024)
            ]))), 128))
          ], 64)) : (u(), h(t(M), {
            key: 0,
            type: "info"
          }, {
            default: e(() => [
              n(a(t(o)("notAvailable")) + " ", 1),
              F("strong", null, a(_.value.gathering.status), 1),
              s[0] || (s[0] = n(". ", -1)),
              _.value.gathering.status !== "tallied" ? (u(), g(w, { key: 0 }, [
                n(a(t(o)("willBePublished")), 1)
              ], 64)) : B("", !0)
            ]),
            _: 1
          }))
        ], 64)) : B("", !0)
      ]),
      _: 1
    }, 8, ["show"]));
  }
}), xt = /* @__PURE__ */ at(bt, [["__scopeId", "data-v-140a0d46"]]);
async function X(S) {
  return (await S.json().catch(() => ({ error: S.statusText }))).error ?? `Request failed (${S.status})`;
}
function kt(S, o = "") {
  const k = o.replace(/\/$/, "");
  return {
    async getContext() {
      const v = await fetch(`${k}/v1/api/member/gatherings/${S}`);
      if (!v.ok) throw new L(await X(v), v.status);
      return v.json();
    },
    async submitBallot(v) {
      const y = await fetch(`${k}/v1/api/member/gatherings/${S}/ballot`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ballot_content: v })
      });
      if (!y.ok) throw new L(await X(y), y.status);
      return y.json();
    }
  };
}
export {
  ft as BallotForm,
  L as HttpError,
  xt as VotingResultsWidget,
  ht as VotingWidget,
  kt as createMemberVotingService
};
