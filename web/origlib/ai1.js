/*!
 * HyperMD, copyright (c) by laobubu
 * Distributed under an MIT license: http://laobubu.net/HyperMD/LICENSE
 *
 * Break the Wall between writing and preview, in a Markdown Editor.
 *
 * HyperMD makes Markdown editor on web WYSIWYG, based on CodeMirror
 *
 * Homepage: http://laobubu.net/HyperMD/
 * Issues: https://github.com/laobubu/HyperMD/issues
 */
!function (e, t) {
    "object" == typeof exports && "undefined" != typeof module ? t(exports, require("codemirror"), require("codemirror/addon/fold/foldcode"), require("codemirror/addon/fold/foldgutter"), require("codemirror/addon/fold/markdown-fold"), require("codemirror/addon/edit/closebrackets"), require("codemirror/lib/codemirror.css"), require("codemirror/addon/fold/foldgutter.css"), require("./theme/hypermd-light.css"), require("codemirror/mode/markdown/markdown"), require("./mode/hypermd.css"), require("codemirror/mode/meta")) : "function" == typeof define && define.amd ? define(["exports", "codemirror", "codemirror/addon/fold/foldcode", "codemirror/addon/fold/foldgutter", "codemirror/addon/fold/markdown-fold", "codemirror/addon/edit/closebrackets", "codemirror/lib/codemirror.css", "codemirror/addon/fold/foldgutter.css", "./theme/hypermd-light.css", "codemirror/mode/markdown/markdown", "./mode/hypermd.css", "codemirror/mode/meta"], t) : t(e.HyperMD = {}, e.CodeMirror);
}(this, function (e, U) {
    "use strict";
    "function" != typeof Object.assign && Object.defineProperty(Object, "assign", {
        value: function (e, t) {
            var n = arguments;
            if (null == e) throw new TypeError("Cannot convert undefined or null to object");
            for (var r = Object(e), i = 1; i < arguments.length; i++) {
                var o = n[i];
                if (null != o) for (var a in o) Object.prototype.hasOwnProperty.call(o, a) && (r[a] = o[a]);
            }
            return r;
        }, writable: !0, configurable: !0
    });
    var s = function (e, t, n, r) {
        void 0 === n && (n = !1), void 0 === r && (r = "enabled"), this.on_cb = e, this.off_cb = t, this.state = n, this.subkey = r;
    };

    function d(t, n, r) {
        n = ~~n || 5;
        var i = 250;
        setTimeout(function e() {
            if (n--) {
                try {
                    if (t()) return;
                } catch (e) {
                }
                setTimeout(e, i), i *= 2;
            } else r && r();
        }, 0);
    }

    function h(e, t) {
        var n = null, r = 0, i = function () {
            e(), n = 0;
        }, o = function () {
            var e = +new Date;
            if (n) {
                if (e < r) return;
                clearTimeout(n);
            }
            n = setTimeout(i, t), r = e + 100;
        };
        return o.stop = function () {
            n && (clearTimeout(n), n = 0);
        }, o;
    }

    s.prototype.ON = function (e) {
        return this.on_cb = e, this;
    }, s.prototype.OFF = function (e) {
        return this.off_cb = e, this;
    }, s.prototype.set = function (e, t) {
        var n = "object" == typeof e && e ? e[this.subkey] : e;
        t && (n = !!n), n !== this.state && ((this.state = n) ? this.on_cb && this.on_cb(n) : this.off_cb && this.off_cb(n));
    }, s.prototype.setBool = function (e) {
        return this.set(e, !0);
    }, s.prototype.bind = function (e, t, n) {
        var r = this;
        return Object.defineProperty(e, t, {
            get: function () {
                return r.state;
            }, set: function (e) {
                return r.set(e, n);
            }, configurable: !0, enumerable: !0
        }), this;
    };
    var t = U.addClass, n = U.rmClass, r = U.contains;

    function a(e, t) {
        var n = new Array(t);
        if (n.fill) n.fill(e); else for (var r = 0; r < t; r++) n[r] = e;
        return n;
    }

    function $(e, t) {
        for (var n = ""; 0 < t--;) n += e;
        return n;
    }

    function m(e, t) {
        for (var n, r = [e]; n = r.shift();) for (var i = 0; i < n.length; i++) {
            var o = n[i];
            o && o.nodeType == Node.ELEMENT_NODE && (t(o), o.children && 0 < o.children.length && r.push(o.children));
        }
    }

    function p(r, i, o) {
        var e = r.getBoundingClientRect(), a = e.width, l = e.height, s = h(function () {
            var e = r.getBoundingClientRect(), t = e.width, n = e.height;
            a == t && l == n || (i(t, n, a, l), a = t, l = n, setTimeout(s, 200));
        }, 100), t = null;
        var n = !1;
        var c = [];
        return o || m([r], function (e) {
            var t, n = e.tagName, r = getComputedStyle(e);
            "none" != (t = "resize", r.getPropertyValue(t) || "") && (o = !0), /^(?:img|video)$/i.test(n) ? (e.addEventListener("load", s, !1), e.addEventListener("error", s, !1)) : /^(?:details|summary)$/i.test(n) && e.addEventListener("click", s, !1);
        }), o && (t = setTimeout(function e() {
            t && clearTimeout(t), n || (t = setTimeout(e, 200)), s();
        }, 200)), {
            check: s, stop: function () {
                n = !0, s.stop(), t && (clearTimeout(t), t = null);
                for (var e = 0; e < c.length; e++) c[e][0].removeEventListener(c[e][1], s, !1);
            }
        };
    }

    function i(e) {
        return "function" == typeof Symbol ? Symbol(e) : "_\n" + e + "\n_" + Math.floor(65535 * Math.random()).toString(16);
    }

    var o = "__hypermd__", l = {
        lineNumbers: !0,
        lineWrapping: !0,
        theme: "hypermd-light",
        mode: "text/x-hypermd",
        tabSize: 4,
        autoCloseBrackets: !0,
        foldGutter: !0,
        gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter", "HyperMD-goback"]
    }, c = {theme: "default"};

    function u(e, t) {
        if (t >= e.display.viewTo) return null;
        if ((t -= e.display.viewFrom) < 0) return null;
        for (var n = e.display.view, r = 0; r < n.length; r++) if ((t -= n[r].size) < 0) return r;
    }

    function b(e, t) {
        if (t >= e.display.viewFrom && t < e.display.viewTo) return e.display.view[u(e, t)];
        var n = e.display.externalMeasured;
        return n && t >= n.lineN && t < n.lineN + n.size ? n : void 0;
    }

    function _(e, t, n) {
        if (e.line == t) return {map: e.measure.map, cache: e.measure.cache, before: !1};
        for (var r = 0; r < e.rest.length; r++) if (e.rest[r] == t) return {
            map: e.measure.maps[r],
            cache: e.measure.caches[r],
            before: !1
        };
        for (var i = 0; i < e.rest.length; i++) if (e.rest[i].lineNo() > n) return {
            map: e.measure.maps[i],
            cache: e.measure.caches[i],
            before: !0
        };
    }

    var f = Object.freeze({findViewIndex: u, findViewForLine: b, mapFromLineView: _}), z = function (e) {
        this.cm = e;
    };

    function w(e, t, n) {
        var r, i = t.line, o = {line: i, ch: 0}, a = {line: i, ch: t.ch}, l = "function" == typeof n && n,
            s = !l && new RegExp("(?:^|\\s)" + n + "(?:\\s|$)"), c = e.getLineTokens(i);
        for (r = 0; r < c.length && !(c[r].end >= t.ch); r++) ;
        if (r === c.length) return null;
        for (var d = r; d < c.length; d++) {
            var h = c[d];
            if (l ? !l(h) : !s.test(h.type)) break;
            a.ch = h.end;
        }
        for (d = r; 0 <= d; d--) {
            h = c[d];
            if (!(l ? l(h) : s.test(h.type))) {
                o.ch = h.end;
                break;
            }
        }
        return {from: o, to: a};
    }

    function g(e) {
        return "anchor" in e && (e = [e.head, e.anchor]), 0 < U.cmpPos(e[0], e[1]) ? [e[1], e[0]] : [e[0], e[1]];
    }

    function k(e, t) {
        var n = e[0], r = e[1], i = t[0], o = t[1];
        return !(U.cmpPos(r, i) < 0 || 0 < U.cmpPos(n, o));
    }

    z.prototype.findNext = function (n, e, r) {
        var i = this.lineNo, o = this.lineTokens, a = null, l = this.i_token + 1, t = !1;
        if (!0 === e ? t = !0 : "number" == typeof e && (l = e), r) if (r.line > i) l = o.length; else if (r.line < i) ; else for (; l < o.length && !(o[l].start >= r.ch); l++) ;
        for (; l < o.length; l++) {
            var s = o[l];
            if ("function" == typeof n ? n(s, o, l) : n.test(s.type)) {
                a = s;
                break;
            }
        }
        if (!a && t) {
            var c = this.cm, d = Math.max(r ? r.line : 0, i + 1);
            c.eachLine(d, c.lastLine() + 1, function (e) {
                if (i = e.lineNo(), o = c.getLineTokens(i), l = 0, r && i === r.line) for (; l < o.length && !(o[l].start >= r.ch); l++) ;
                for (; l < o.length; l++) {
                    var t = o[l];
                    if ("function" == typeof n ? n(t, o, l) : n.test(t.type)) return a = t, !0;
                }
            });
        }
        return a ? {lineNo: i, token: a, i_token: l} : null;
    }, z.prototype.findPrev = function (e, t, n) {
        var r = this.lineNo, i = this.lineTokens, o = null, a = this.i_token - 1, l = !1;
        if (!0 === t ? l = !0 : "number" == typeof t && (a = t), n) if (n.line < r) a = -1; else if (n.line > r) ; else for (; a < i.length && !(i[a].start >= n.ch); a++) ;
        for (a >= i.length && (a = i.length - 1); 0 <= a; a--) {
            var s = i[a];
            if ("function" == typeof e ? e(s, i, a) : e.test(s.type)) {
                o = s;
                break;
            }
        }
        if (!o && l) {
            var c = this.cm, d = Math.min(n ? n.line : c.lastLine(), r - 1), h = c.firstLine();
            for (r = d + 1; !o && h <= --r;) {
                c.getLineHandle(r);
                if (i = c.getLineTokens(r), a = 0, n && r === n.line) for (; a < i.length && !(i[a].start >= n.ch); a++) ;
                for (a >= i.length && (a = i.length - 1); 0 <= a; a--) {
                    s = i[a];
                    if ("function" == typeof e ? e(s, i, a) : e.test(s.type)) {
                        o = s;
                        break;
                    }
                }
            }
        }
        return o ? {lineNo: r, token: o, i_token: a} : null;
    }, z.prototype.expandRange = function (t, e) {
        var n, r = this.cm;
        n = "function" == typeof t ? t : ("string" == typeof t && (t = new RegExp("(?:^|\\s)" + t + "(?:\\s|$)")), function (e) {
            return !!e && t.test(e.type || "");
        });
        for (var i = {
            lineNo: this.lineNo,
            i_token: this.i_token,
            token: this.lineTokens[this.i_token]
        }, o = Object.assign({}, i), a = !1, l = this.lineTokens, s = this.i_token; !a;) {
            for (s >= l.length && (s = l.length - 1); 0 <= s; s--) {
                var c = l[s];
                if (!n(c, l, s)) {
                    a = !0;
                    break;
                }
                i.i_token = s, i.token = c;
            }
            if (a || !(e && i.lineNo > r.firstLine())) break;
            s = (l = r.getLineTokens(--i.lineNo)).length - 1;
        }
        for (a = !1, l = this.lineTokens, s = this.i_token; !a;) {
            for (s < 0 && (s = 0); s < l.length; s++) {
                var d = l[s];
                if (!n(d, l, s)) {
                    a = !0;
                    break;
                }
                o.i_token = s, o.token = d;
            }
            if (a || !(e && o.lineNo < r.lastLine())) break;
            l = r.getLineTokens(++o.lineNo), s = 0;
        }
        return {from: i, to: o};
    }, z.prototype.setPos = function (e, t, n) {
        void 0 === t ? (t = e, e = this.line) : "number" == typeof e && (e = this.cm.getLineHandle(e));
        var r = e === this.line, i = 0;
        n || !r ? (this.line = e, this.lineNo = e.lineNo(), this.lineTokens = this.cm.getLineTokens(this.lineNo)) : (i = this.i_token, this.lineTokens[i].start > t && (i = 0));
        for (var o = this.lineTokens; i < o.length && !(o[i].end > t); i++) ;
        this.i_token = i;
    }, z.prototype.getToken = function (e) {
        return "number" != typeof e && (e = this.i_token), this.lineTokens[e];
    }, z.prototype.getTokenType = function (e) {
        "number" != typeof e && (e = this.i_token);
        var t = this.lineTokens[e];
        return t && t.type || "";
    };
    var v = function (e) {
        var r = this;
        this.cm = e, this.caches = new Array, e.on("change", function (e, t) {
            var n = t.from.line;
            r.caches.length > n && r.caches.splice(n);
        });
    };
    v.prototype.getTokenTypes = function (e, t) {
        var n = t ? t.state : {}, r = e.state, i = " " + e.type + " ";
        return {
            em: r.em ? 1 : n.em ? 2 : 0,
            strikethrough: r.strikethrough ? 1 : n.strikethrough ? 2 : 0,
            strong: r.strong ? 1 : n.strong ? 2 : 0,
            code: r.code ? 1 : n.code ? 2 : 0,
            linkText: r.linkText ? 3 === r.hmdLinkType || 6 === r.hmdLinkType ? 1 : 0 : n.linkText ? 2 : 0,
            linkHref: r.linkHref && !r.linkText ? 1 : r.linkHref || r.linkText || !n.linkHref || n.linkText ? 0 : 2,
            task: -1 !== i.indexOf(" formatting-task ") ? 3 : 0,
            hashtag: r.hmdHashtag ? 1 : n.hmdHashtag ? 2 : 0
        };
    }, v.prototype.extract = function (e, t) {
        if (!t) {
            var n = this.caches[e];
            if (n) return n;
        }
        for (var r = this.cm.getLineTokens(e), i = this.cm.getLine(e), o = i.length, a = [], l = {}, s = 0; s < r.length; s++) {
            var c = r[s], d = this.getTokenTypes(c, r[s - 1]);
            for (var h in d) {
                var u = l[h];
                1 & d[h] && (u || (u = {
                    type: h,
                    begin: c.start,
                    end: o,
                    head: c,
                    head_i: s,
                    tail: r[r.length - 1],
                    tail_i: r.length - 1,
                    text: i.slice(c.start)
                }, a.push(u), l[h] = u)), 2 & d[h] && u && (u.tail = c, u.tail_i = s, u.end = c.end, u.text = u.text.slice(0, u.end - u.begin), l[h] = null);
            }
        }
        return this.caches[e] = a;
    }, v.prototype.findSpansAt = function (e) {
        for (var t = this.extract(e.line), n = e.ch, r = [], i = 0; i < t.length; i++) {
            var o = t[i];
            if (o.begin > n) break;
            n >= o.begin && o.end >= n && r.push(o);
        }
        return r;
    }, v.prototype.findSpanWithTypeAt = function (e, t) {
        for (var n = this.extract(e.line), r = e.ch, i = 0; i < n.length; i++) {
            var o = n[i];
            if (o.begin > r) break;
            if (r >= o.begin && o.end >= r && o.type === t) return o;
        }
        return null;
    };
    var y = i("LineSpanExtractor");

    function T(e) {
        return y in e ? e[y] : e[y] = new v(e);
    }

    function x(r, i, o) {
        return function (e) {
            if (e.hmd || (e.hmd = {}), e.hmd[r]) return e.hmd[r];
            var t = new i(e);
            if (e.hmd[r] = t, o) for (var n in o) t[n] = o[n];
            return t;
        };
    }

    var L = Object.freeze({
            Addon: function (e) {
            }, Getter: x
        }), W = /[^\\][$|]/, B = /^(?:[*\-+]|^[0-9]+([.)]))\s+/,
        X = /^((?:(?:aaas?|about|acap|adiumxtra|af[ps]|aim|apt|attachment|aw|beshare|bitcoin|bolo|callto|cap|chrome(?:-extension)?|cid|coap|com-eventbrite-attendee|content|crid|cvs|data|dav|dict|dlna-(?:playcontainer|playsingle)|dns|doi|dtn|dvb|ed2k|facetime|feed|file|finger|fish|ftp|geo|gg|git|gizmoproject|go|gopher|gtalk|h323|hcp|https?|iax|icap|icon|im|imap|info|ipn|ipp|irc[6s]?|iris(?:\.beep|\.lwz|\.xpc|\.xpcs)?|itms|jar|javascript|jms|keyparc|lastfm|ldaps?|magnet|mailto|maps|market|message|mid|mms|ms-help|msnim|msrps?|mtqp|mumble|mupdate|mvn|news|nfs|nih?|nntp|notes|oid|opaquelocktoken|palm|paparazzi|platform|pop|pres|proxy|psyc|query|res(?:ource)?|rmi|rsync|rtmp|rtsp|secondlife|service|session|sftp|sgn|shttp|sieve|sips?|skype|sm[bs]|snmp|soap\.beeps?|soldat|spotify|ssh|steam|svn|tag|teamspeak|tel(?:net)?|tftp|things|thismessage|tip|tn3270|tv|udp|unreal|urn|ut2004|vemmi|ventrilo|view-source|webcal|wss?|wtai|wyciwyg|xcon(?:-userid)?|xfire|xmlrpc\.beeps?|xmpp|xri|ymsgr|z39\.50[rs]?):(?:\/{1,3}|[a-z0-9%])|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}\/)(?:[^\s()<>]|\([^\s()<>]*\))+(?:\([^\s()<>]*\)|[^\s`*!()\[\]{};:'".,<>?«»“”‘’]))/i,
        Y = /^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/,
        G = /^\.{0,2}\/[^\>\s]+/,
        V = /^(?:[-()/a-zA-Z0-9ァ-ヺー-ヾｦ-ﾟｰ０-９Ａ-Ｚａ-ｚぁ-ゖ゙-ゞー々ぁ-んァ-ヾ一-\u9FEF㐀-䶵﨎﨏﨑﨓﨔﨟﨡﨣﨤﨧-﨩]|[\ud840-\ud868][\udc00-\udfff]|\ud869[\udc00-\uded6\udf00-\udfff]|[\ud86a-\ud86c][\udc00-\udfff]|\ud86d[\udc00-\udf34\udf40-\udfff]|\ud86e[\udc00-\udc1d])+/,
        Z = /^\s*[^\|].*?\|.*[^|]\s*$/, Q = /^\s*[^\|].*\|/, J = /^\s*\|[^\|]+\|.+\|\s*$/, ee = /^\s*\|/, te = {};

    function ne(e) {
        e.hmdTable = 0, e.hmdTableColumns = [], e.hmdTableID = null, e.hmdTableCol = e.hmdTableRow = 0;
    }

    te[1] = "hmd-barelink", te[6] = "hmd-barelink2", te[2] = "hmd-barelink hmd-footref", te[4] = "hmd-footnote line-HyperMD-footnote", te[7] = "hmd-footref2";
    var re = /^\s+((\d+[).]|[-*+])\s+)?/, M = {
        hr: "line-HyperMD-hr line-background-HyperMD-hr-bg hr",
        list1: "list-1",
        list2: "list-2",
        list3: "list-3",
        code: "inline-code",
        hashtag: "hashtag meta"
    };
    U.defineMode("hypermd", function (j, e) {
        var F = {
            front_matter: !0,
            math: !0,
            table: !0,
            toc: !0,
            orgModeMarkup: !0,
            hashtag: !1,
            fencedCodeBlockHighlighting: !0,
            name: "markdown",
            highlightFormatting: !0,
            taskLists: !0,
            strikethrough: !0,
            emoji: !0,
            tokenTypeOverrides: M
        };
        Object.assign(F, e), F.tokenTypeOverrides !== M && (F.tokenTypeOverrides = Object.assign({}, M, F.tokenTypeOverrides)), F.name = "markdown";
        var P = {htmlBlock: null}, q = U.getMode(j, F), t = Object.assign({}, q);

        function o(e, t) {
            var n = t.hmdInnerExitChecker(e, t), r = t.hmdInnerStyle,
                i = (!n || !n.skipInnerMode) && t.hmdInnerMode.token(e, t.hmdInnerState) || "";
            return r && (i += " " + r), n && (n.style && (i += " " + n.style), n.endPos && (e.pos = n.endPos), t.hmdInnerExitChecker = null, t.hmdInnerMode = null, t.hmdInnerState = null, t.hmdOverride = null), i.trim() || null;
        }

        function $(n) {
            return {
                token: function (e) {
                    var t = e.string.indexOf(n, e.start);
                    return -1 === t ? e.skipToEnd() : 0 === t ? e.pos += n.length : (e.pos = t, "\\" === e.string.charAt(t - 1) && e.pos++), null;
                }
            };
        }

        function z(n, r) {
            return r || (r = {}), function (e, t) {
                return e.string.substr(e.start, n.length) === n ? (r.endPos = e.start + n.length, r) : null;
            };
        }

        function K(e, t, n, r) {
            if ("string" == typeof n && (n = U.getMode(j, n)), !(n && "null" !== n.name || (n = "endTag" in r ? $(r.endTag) : "function" == typeof r.fallbackMode && r.fallbackMode()))) throw new Error("no mode");
            t.hmdInnerExitChecker = "endTag" in r ? z(r.endTag) : r.exitChecker, t.hmdInnerStyle = r.style, t.hmdInnerMode = n, t.hmdOverride = o, t.hmdInnerState = U.startState(n);
            var i = r.style || "";
            return r.skipFirstToken || (i += " " + n.token(e, t.hmdInnerState)), i.trim();
        }

        return t.startState = function () {
            var e = q.startState();
            return ne(e), e.hmdOverride = null, e.hmdInnerExitChecker = null, e.hmdInnerMode = null, e.hmdLinkType = 0, e.hmdNextMaybe = F.front_matter ? 1 : 0, e.hmdNextState = null, e.hmdNextStyle = null, e.hmdNextPos = null, e.hmdHashtag = 0, e;
        }, t.copyState = function (e) {
            for (var t = q.copyState(e), n = 0, r = ["hmdLinkType", "hmdNextMaybe", "hmdTable", "hmdTableID", "hmdTableCol", "hmdTableRow", "hmdOverride", "hmdInnerMode", "hmdInnerStyle", "hmdInnerExitChecker", "hmdNextPos", "hmdNextState", "hmdNextStyle", "hmdHashtag"]; n < r.length; n += 1) {
                var i = r[n];
                t[i] = e[i];
            }
            return t.hmdTableColumns = e.hmdTableColumns.slice(0), e.hmdInnerMode && (t.hmdInnerState = U.copyState(e.hmdInnerMode, e.hmdInnerState)), t;
        }, t.blankLine = function (e) {
            var t, n = e.hmdInnerMode;
            return n ? n.blankLine && (t = n.blankLine(e.hmdInnerState)) : t = q.blankLine(e), t || (t = ""), -1 === e.code && (t += " line-HyperMD-codeblock line-background-HyperMD-codeblock-bg"), ne(e), t.trim() || null;
        }, t.indent = function (e, t) {
            var n = e.hmdInnerMode || q, r = n.indent;
            return "function" == typeof r ? r.apply(n, arguments) : U.Pass;
        }, t.innerMode = function (e) {
            return e.hmdInnerMode ? {mode: e.hmdInnerMode, state: e.hmdInnerState} : q.innerMode(e);
        }, t.token = function (e, t) {
            if (t.hmdOverride) return t.hmdOverride(e, t);
            if (1 === t.hmdNextMaybe) {
                if ("---" === e.string) return t.hmdNextMaybe = 2, K(e, t, "yaml", {
                    style: "hmd-frontmatter",
                    fallbackMode: function () {
                        return $("---");
                    },
                    exitChecker: function (e, t) {
                        return "---" === e.string ? (t.hmdNextMaybe = 0, {endPos: 3}) : null;
                    }
                });
                t.hmdNextMaybe = 0;
            }
            var n, r = t.f === P.htmlBlock, i = -1 === t.code, o = 0 === e.start, a = t.linkText, l = t.linkHref,
                s = !(i || r), c = s && !(t.code || t.indentedCode || t.linkHref), d = "";
            if (s) {
                if (F.math && c && (n = e.match(/^\${1,2}/, !1))) {
                    var h = n[0], u = h.length;
                    if (2 === u || e.string.slice(e.pos).match(/[^\\]\$/)) {
                        var f = U.getMode(j, {name: "stex"}), m = "stex" !== f.name;
                        return d += K(e, t, f, {
                            style: "math", skipFirstToken: m, fallbackMode: function () {
                                return $(h);
                            }, exitChecker: z(h, {style: "formatting formatting-math formatting-math-end math-" + u})
                        }), m && (e.pos += n[0].length), d += " formatting formatting-math formatting-math-begin math-" + u;
                    }
                }
                if (o && F.orgModeMarkup && (n = e.match(/^\#\+(\w+\:?)\s*/))) return e.eol() || (t.hmdOverride = function (e, t) {
                    return e.skipToEnd(), t.hmdOverride = null, "string hmd-orgmode-markup";
                }), "meta formatting-hmd-orgmode-markup hmd-orgmode-markup line-HyperMD-orgmode-markup";
                if (o && F.toc && e.match(/^\[TOCM?\]\s*$/i)) return "meta line-HyperMD-toc hmd-toc";
                if (c && !t.hmdLinkType && (e.match(X) || e.match(Y))) return "url";
            }
            t.hmdNextState ? (e.pos = t.hmdNextPos, d += " " + (t.hmdNextStyle || ""), Object.assign(t, t.hmdNextState), t.hmdNextState = null, t.hmdNextStyle = null, t.hmdNextPos = null) : d += " " + (q.token(e, t) || ""), 0 !== t.hmdHashtag && (d += " " + F.tokenTypeOverrides.hashtag), !P.htmlBlock && t.htmlState && (P.htmlBlock = t.f);
            var p = t.f === P.htmlBlock, g = -1 === t.code;
            s = s && !(p || g), (c = c && s && !(t.code || t.indentedCode || t.linkHref)) && (n = e.current().match(W)) && (e.pos = e.start + n.index + 1);
            var v = e.current();
            if (p != r && (p ? (d += " hmd-html-begin", P.htmlBlock = t.f) : d += " hmd-html-end"), (i || g) && (t.localMode && i || (d = d.replace("inline-code", "")), d += " line-HyperMD-codeblock line-background-HyperMD-codeblock-bg", g !== i && (g ? i || (d += " line-HyperMD-codeblock-begin line-background-HyperMD-codeblock-begin-bg") : d += " line-HyperMD-codeblock-end line-background-HyperMD-codeblock-end-bg")), s) {
                var b = t.hmdTable;
                if (o && b) (1 == b ? Q : ee).test(e.string) ? (t.hmdTableCol = 0, t.hmdTableRow++) : ne(t);
                if (o && t.header && (/^(?:---+|===+)\s*$/.test(e.string) && t.prevLine && t.prevLine.header ? d += " line-HyperMD-header-line line-HyperMD-header-line-" + t.header : d += " line-HyperMD-header line-HyperMD-header-" + t.header), t.indentedCode && (d += " hmd-indented-code"), t.quote) {
                    if (e.eol() && (d += " line-HyperMD-quote line-HyperMD-quote-" + t.quote, /^ {0,3}\>/.test(e.string) || (d += " line-HyperMD-quote-lazy")), o && (n = v.match(/^\s+/))) return e.pos = n[0].length, (d += " hmd-indent-in-quote").trim();
                    /^>\s+$/.test(v) && ">" != e.peek() && (e.pos = e.start + 1, v = ">", t.hmdOverride = function (e, t) {
                        return e.match(re), t.hmdOverride = null, "hmd-indent-in-quote line-HyperMD-quote line-HyperMD-quote-" + t.quote;
                    });
                }
                var _ = (t.listStack[t.listStack.length - 1] || 0) + 3,
                    k = o && /^\s+$/.test(v) && (!1 !== t.list || e.indentation() <= _),
                    y = t.list && /formatting-list/.test(d);
                if (y || k && (!1 !== t.list || e.match(B, !1))) {
                    var w = t.listStack && t.listStack.length || 0;
                    if (k) {
                        if (e.match(B, !1)) !1 === t.list && w++; else {
                            for (; 0 < w && e.pos < t.listStack[w - 1];) w--;
                            if (!w) return d.trim() || null;
                            d += " line-HyperMD-list-line-nobullet line-HyperMD-list-line line-HyperMD-list-line-" + w;
                        }
                        d += " hmd-list-indent hmd-list-indent-" + w;
                    } else y && (d += " line-HyperMD-list-line line-HyperMD-list-line-" + w);
                }
                if (a !== t.linkText && (a ? (t.hmdLinkType in te && (d += " " + te[t.hmdLinkType]), 4 === t.hmdLinkType ? t.hmdLinkType = 5 : t.hmdLinkType = 0) : (n = e.match(/^([^\]]+)\](\(| ?\[|\:)?/, !1)) ? n[2] ? ":" === n[2] ? t.hmdLinkType = 4 : "[" !== n[2] && " [" !== n[2] || "]" !== e.string.charAt(e.pos + n[0].length) ? t.hmdLinkType = 3 : t.hmdLinkType = 6 : "^" === n[1].charAt(0) ? t.hmdLinkType = 2 : t.hmdLinkType = 1 : t.hmdLinkType = 1), l !== t.linkHref && (l ? t.hmdLinkType && (d += " " + te[t.hmdLinkType], t.hmdLinkType = 0) : "[" === v && "]" !== e.peek() && (t.hmdLinkType = 7)), 0 !== t.hmdLinkType && (t.hmdLinkType in te && (d += " " + te[t.hmdLinkType]), 5 === t.hmdLinkType && (/^(?:\]\:)?\s*$/.test(v) || (X.test(v) || G.test(v) ? d += " hmd-footnote-url" : d = d.replace("string url", ""), t.hmdLinkType = 0))), /formatting-escape/.test(d) && 1 < v.length) {
                    var T = v.length - 1, x = d.replace("formatting-escape", "escape") + " hmd-escape-char";
                    return t.hmdOverride = function (e, t) {
                        return e.pos += T, t.hmdOverride = null, x.trim();
                    }, d += " hmd-escape-backslash", e.pos -= T, d;
                }
                if (!d.trim() && F.table) {
                    var L = !1;
                    if ("|" === v.charAt(0) && (e.pos = e.start + 1, v = "|", L = !0), L) {
                        if (!b) {
                            var M;
                            if (Z.test(e.string) ? b = 1 : J.test(e.string) && (b = 2), b) {
                                var C = e.lookAhead(1);
                                if (2 === b ? J.test(C) ? C = C.replace(/^\s*\|/, "").replace(/\|\s*$/, "") : b = 0 : 1 === b && (Z.test(C) || (b = 0)), b) {
                                    M = C.split("|");
                                    for (var O = 0; O < M.length; O++) {
                                        var H = M[O];
                                        if (/^\s*--+\s*:\s*$/.test(H)) H = "right"; else if (/^\s*:\s*--+\s*$/.test(H)) H = "left"; else if (/^\s*:\s*--+\s*:\s*$/.test(H)) H = "center"; else {
                                            if (!/^\s*--+\s*$/.test(H)) {
                                                b = 0;
                                                break;
                                            }
                                            H = "default";
                                        }
                                        M[O] = H;
                                    }
                                }
                            }
                            b && (t.hmdTable = b, t.hmdTableColumns = M, t.hmdTableID = "T" + e.lineOracle.line, t.hmdTableRow = t.hmdTableCol = 0);
                        }
                        if (b) {
                            var S = t.hmdTableColumns.length - 1;
                            if (2 === b && (0 === t.hmdTableCol && /^\s*\|$/.test(e.string.slice(0, e.pos)) || e.match(/^\s*$/, !1))) d += " hmd-table-sep hmd-table-sep-dummy"; else if (t.hmdTableCol < S) {
                                var E = t.hmdTableRow, D = t.hmdTableCol++;
                                0 == D && (d += " line-HyperMD-table_" + t.hmdTableID + " line-HyperMD-table-" + b + " line-HyperMD-table-row line-HyperMD-table-row-" + E), d += " hmd-table-sep hmd-table-sep-" + D;
                            }
                        }
                    }
                }
                if (b && 1 === t.hmdTableRow && /emoji/.test(d) && (d = ""), c && "<" === v) {
                    var N = null;
                    if (e.match(/^\![A-Z]+/) ? N = ">" : e.match("?") ? N = "?>" : e.match("![CDATA[") && (N = "]]>"), null != N) return K(e, t, null, {
                        endTag: N,
                        style: (d + " comment hmd-cdata-html").trim()
                    });
                }
                if (F.hashtag && c) switch (t.hmdHashtag) {
                    case 0:
                        if ("#" === v && !t.linkText && !t.image && (o || /^\s*$/.test(e.string.charAt(e.start - 1)))) {
                            var A = e.string.slice(e.pos).replace(/\\./g, "");
                            if (n = V.exec(A)) for (/^\d+$/.test(n[0]) ? t.hmdHashtag = 0 : t.hmdHashtag = 1, A = A.slice(n[0].length); ;) {
                                if ("#" === A[0] && (1 === A.length || !V.test(A[1]))) {
                                    t.hmdHashtag = 2;
                                    break;
                                }
                                if (!(n = A.match(/^\s+/)) || !(n = (A = A.slice(n[0].length)).match(V))) break;
                                A = A.slice(n[0].length);
                            }
                            t.hmdHashtag && (d += " formatting formatting-hashtag hashtag-begin " + F.tokenTypeOverrides.hashtag);
                        }
                        break;
                    case 1:
                        var R = !1;
                        if (!/formatting/.test(d) && !/^\s*$/.test(v)) {
                            n = v.match(V);
                            var I = v.length - (n ? n[0].length : 0);
                            0 < I && (e.backUp(I), R = !0);
                        }
                        R || (R = e.eol()), R || (R = !V.test(e.peek())), R && (d += " hashtag-end", t.hmdHashtag = 0);
                        break;
                    case 2:
                        "#" === v && (d = d.replace(/\sformatting-header(?:-\d+)?/g, ""), d += " formatting formatting-hashtag hashtag-end", t.hmdHashtag = 0);
                }
            }
            return d.trim() || null;
        }, t;
    }, "hypermd"), U.defineMIME("text/x-hypermd", "hypermd");
    var C = Object.freeze({});
    var O = {byDrop: !1, byPaste: !1, fileHandler: null}, H = {byPaste: !0, byDrop: !0};
    l.hmdInsertFile = H, U.defineOption("hmdInsertFile", O, function (e, t) {
        if (t && "boolean" != typeof t) "function" == typeof t && (t = {
            byDrop: !0,
            byPaste: !0,
            fileHandler: t
        }); else {
            var n = !!t;
            t = {byDrop: n, byPaste: n};
        }
        var r = E(e);
        for (var i in O) r[i] = i in t ? t[i] : O[i];
    });
    var S = function (e) {
        var o = this;
        this.cm = e, this.pasteHandle = function (e, t) {
            o.doInsert(t.clipboardData || window.clipboardData, !0) && t.preventDefault();
        }, this.dropHandle = function (t, n) {
            var r = o, i = (t = o.cm, !1);
            t.operation(function () {
                var e = t.coordsChar({left: n.clientX, top: n.clientY}, "window");
                t.setCursor(e), i = r.doInsert(n.dataTransfer, !1);
            }), i && n.preventDefault();
        }, new s(function () {
            return o.cm.on("paste", o.pasteHandle);
        }, function () {
            return o.cm.off("paste", o.pasteHandle);
        }).bind(this, "byPaste", !0), new s(function () {
            return o.cm.on("drop", o.dropHandle);
        }, function () {
            return o.cm.off("drop", o.dropHandle);
        }).bind(this, "byDrop", !0);
    };
    S.prototype.doInsert = function (e, t) {
        var a = this.cm;
        if (t && e.types && e.types.some(function (e) {
            return "text/" === e.slice(0, 5);
        })) return !1;
        if (!e || !e.files || 0 === e.files.length) return !1;
        var r = e.files, i = this.fileHandler, l = !1;
        return "function" == typeof i && (a.operation(function () {
            a.replaceSelection(".");
            var e = a.getCursor(), t = {line: e.line, ch: e.ch - 1}, n = document.createElement("span"),
                o = a.markText(t, e, {replacedWith: n, clearOnEnter: !1, handleMouseEvents: !1});
            (l = i(r, {
                marker: o, cm: a, finish: function (r, i) {
                    return a.operation(function () {
                        var e = o.find(), t = e.from, n = e.to;
                        a.replaceRange(r, t, n), o.clear(), "number" == typeof i && a.setCursor({
                            line: t.line,
                            ch: t.ch + i
                        });
                    });
                }, setPlaceholder: function (e) {
                    0 < n.childNodes.length && n.removeChild(n.firstChild), n.appendChild(e), o.changed();
                }, resize: function () {
                    o.changed();
                }
            })) || o.clear();
        }), l);
    };
    var E = x("InsertFile", S, O), D = Object.freeze({
        ajaxUpload: function (e, t, n, r) {
            var i = new XMLHttpRequest, o = new FormData;
            for (var a in t) o.append(a, t[a]);
            i.onreadystatechange = function () {
                if (4 == this.readyState) {
                    var e = i.responseText;
                    try {
                        e = JSON.parse(i.responseText);
                    } catch (e) {
                    }
                    /^20\d/.test(i.status + "") ? n(e, null) : n(null, e);
                }
            }, i.open(r || "POST", e, !0), i.send(o);
        }, defaultOption: O, suggestedOption: H, InsertFile: S, getAddon: E
    });

    function N(e) {
        var t = e = e.trim(), n = "", r = e.match(/^(\S+)\s+("(?:[^"\\]+|\\.)+"|[^"\s].*)/);
        return r && (t = r[1], '"' === (n = r[2]).charAt(0) && (n = n.substr(1, n.length - 2).replace(/\\"/g, '"'))), {
            url: t,
            title: n
        };
    }

    var A = {
        hmdReadLink: function (e, t) {
            return P(this).read(e, t);
        }, hmdResolveURL: function (e, t) {
            return P(this).resolve(e, t);
        }, hmdSplitLink: N
    };
    for (var R in A) U.defineExtension(R, A[R]);
    var I = {baseURI: ""}, j = {baseURI: ""};
    l.hmdReadLink = j, U.defineOption("hmdReadLink", I, function (e, t) {
        t && "string" != typeof t || (t = {baseURI: t});
        var n = P(e);
        for (var r in I) n[r] = r in t ? t[r] : I[r];
    });
    var F = function (e) {
        var t = this;
        this.cm = e, this.cache = {}, e.on("changes", h(function () {
            return t.rescan();
        }, 500)), this.rescan();
    };
    F.prototype.read = function (e, t) {
        var n, r = this.cache[e.trim().toLowerCase()] || [];
        "number" != typeof t && (t = 1e9);
        for (var i = 0; i < r.length && !((n = r[i]).line > t); i++) ;
        return n;
    }, F.prototype.rescan = function () {
        var e = this.cm, o = this.cache = {};
        e.eachLine(function (e) {
            var t = e.text, n = /^(?:>\s+)*>?\s{0,3}\[([^\]]+)\]:\s*(.+)$/.exec(t);
            if (n) {
                var r = n[1].trim().toLowerCase(), i = n[2];
                o[r] || (o[r] = []), o[r].push({line: e.lineNo(), content: i});
            }
        });
    }, F.prototype.resolve = function (e, t) {
        var n, r = /^(?:[\w-]+\:\/*|\/\/)[^\/]+/, i = /\/[^\/]+(?:\/+\.?)*$/;
        if (!e) return e;
        if (/^(([^<>()[\]\\.,;:\s@\"]+(\.[^<>()[\]\\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/.test(e)) return "mailto:" + e;
        var o = "";
        if (!(t = t || this.baseURI) || r.test(e)) return e;
        for ((n = t.match(r)) && (o = n[0], t = t.slice(o.length)); n = e.match(/^(\.{1,2})([\/\\]+)/);) e = e.slice(n[0].length), ".." == n[1] && (t = t.replace(i, ""));
        return e = "/" === e.charAt(0) && o ? o + e : (/\/$/.test(t) || (t += "/"), o + t + e);
    };
    var P = x("ReadLink", F, I), q = Object.freeze({
        splitLink: N,
        Extensions: A,
        defaultOption: I,
        suggestedOption: j,
        ReadLink: F,
        getAddon: P
    }), K = "function" == typeof marked ? marked : function (e) {
        return "<pre>" + (e = e.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/  /g, " &nbsp;")) + "</pre>";
    };

    function ie(e, t) {
        return t ? K(t) : null;
    }

    var oe = {enabled: !1, xOffset: 10, convertor: ie}, ae = {enabled: !0};
    l.hmdHover = ae, U.defineOption("hmdHover", oe, function (e, t) {
        t && "boolean" != typeof t ? "function" == typeof t && (t = {enabled: !0, convertor: t}) : t = {enabled: !!t};
        var n = ce(e);
        for (var r in oe) n[r] = r in t ? t[r] : oe[r];
    });
    var le = function (e) {
        var t = this;
        this.cm = e, new s(function () {
            n.addEventListener("mouseenter", a, !0);
        }, function () {
            n.removeEventListener("mouseenter", a, !0), t.hideInfo();
        }).bind(this, "enabled", !0);
        var n = e.display.lineDiv;
        this.lineDiv = n;
        var r = document.createElement("div"), i = document.createElement("div"), o = document.createElement("div");
        r.setAttribute("style", "position:absolute;z-index:99"), r.setAttribute("class", "HyperMD-hover"), r.setAttribute("cm-ignore-events", "true"), i.setAttribute("class", "HyperMD-hover-content"), r.appendChild(i), o.setAttribute("class", "HyperMD-hover-indicator"), r.appendChild(o), this.tooltipDiv = r, this.tooltipContentDiv = i, this.tooltipIndicator = o;
        var a = this.mouseenter.bind(this);
    };
    le.prototype.mouseenter = function (e) {
        var t, n = this.cm, r = e.target, i = r.className;
        if (!(r == this.tooltipDiv || r.compareDocumentPosition && 8 == (8 & r.compareDocumentPosition(this.tooltipDiv)))) if ("SPAN" === r.nodeName && (t = i.match(/(?:^|\s)cm-(hmd-barelink2?|hmd-footref2)(?:\s|$)/))) {
            var o = n.coordsChar({left: e.clientX, top: e.clientY}, "window"), a = null, l = null, s = w(n, o, t[1]);
            s && (a = (a = n.getRange(s.from, s.to)).slice(1, -1)) && (l = n.hmdReadLink(a, o.line) || null);
            var c = (this.convertor || ie)(a, l && l.content || null);
            c ? this.showInfo(c, r) : this.hideInfo();
        } else this.hideInfo();
    }, le.prototype.showInfo = function (e, t) {
        var n = t.getBoundingClientRect(), r = this.lineDiv.getBoundingClientRect(), i = this.tooltipDiv,
            o = this.xOffset;
        this.tooltipContentDiv.innerHTML = e, i.style.left = n.left - r.left - o + "px", this.lineDiv.appendChild(i);
        var a = i.getBoundingClientRect();
        a.right > r.right && (o = a.right - r.right, i.style.left = n.left - r.left - o + "px"), i.style.top = n.top - r.top - a.height + "px", this.tooltipIndicator.style.marginLeft = o + "px";
    }, le.prototype.hideInfo = function () {
        this.tooltipDiv.parentElement == this.lineDiv && this.lineDiv.removeChild(this.tooltipDiv);
    };
    var se, ce = x("Hover", le, oe),
        de = Object.freeze({defaultConvertor: ie, defaultOption: oe, suggestedOption: ae, Hover: le, getAddon: ce}),
        he = function (e, t) {
            var n = e.text, r = e.type, i = e.url, o = e.pos;
            if ("url" === r || "link" === r) {
                var a = n.match(/\[[^\[\]]+\](?:\[\])?$/);
                a && e.altKey ? ("[]" === (n = a[0]).slice(-2) && (n = n.slice(0, -2)), r = "footref") : (e.ctrlKey || e.altKey) && i && window.open(i, "_blank");
            }
            if ("todo" === r) {
                var l = w(t, o, "formatting-task"), s = l.from, c = l.to, d = t.getRange(s, c);
                d = "[ ]" === d ? "[x]" : "[ ]", t.replaceRange(d, s, c);
            }
            if ("footref" === r && (e.ctrlKey || e.altKey)) {
                var h = n.slice(1, -1), u = t.hmdReadLink(h, o.line);
                u && (ue(t, u.line, o), t.setCursor({line: u.line, ch: 0}));
            }
        }, ue = (se = null, function (e, t, n) {
            var r, i, o = function (e) {
                if (-1 == e.options.gutters.indexOf("HyperMD-goback")) return null;
                var t = document.createElement("div");
                t.className = "HyperMD-goback-button", t.addEventListener("click", function () {
                    e.setCursor(se.find()), e.clearGutter("HyperMD-goback"), se.clear(), se = null;
                });
                var n = e.display.gutters.children;
                return n = (n = n[n.length - 1]).offsetLeft + n.offsetWidth, t.style.width = n + "px", t.style.marginLeft = -n + "px", t;
            }(e);
            o && (o.innerHTML = n.line + 1 + "", r = e, i = n, se && (r.clearGutter("HyperMD-goback"), se.clear()), se = r.setBookmark(i), e.setGutterMarker(t, "HyperMD-goback", o));
        }), fe = {enabled: !1, handler: null}, me = {enabled: !0};
    l.hmdClick = me, U.defineOption("hmdClick", fe, function (e, t) {
        t && "boolean" != typeof t ? "function" == typeof t && (t = {enabled: !0, handler: t}) : t = {enabled: !!t};
        var n = ge(e);
        for (var r in fe) n[r] = r in t ? t[r] : fe[r];
    });
    var pe = function (e) {
            var y = this;
            this.cm = e, this._mouseMove_keyDetect = function (e) {
                var t = y.el, n = t.className, r = n, i = "HyperMD-with-alt", o = "HyperMD-with-ctrl";
                !e.altKey && 0 <= n.indexOf(i) && (r = n.replace(i, "")), !e.ctrlKey && 0 <= n.indexOf(o) && (r = n.replace(o, "")), e.altKey || e.ctrlKey || (y._KeyDetectorActive = !1, t.removeEventListener("mousemove", y._mouseMove_keyDetect, !1)), n != r && (t.className = r.trim());
            }, this._keyDown = function (e) {
                var t = e.keyCode || e.which, n = "";
                17 == t && (n = "HyperMD-with-ctrl"), 18 == t && (n = "HyperMD-with-alt");
                var r = y.el;
                n && -1 == r.className.indexOf(n) && (r.className += " " + n), y._KeyDetectorActive || (y._KeyDetectorActive = !0, y.el.addEventListener("mousemove", y._mouseMove_keyDetect, !1));
            }, this._mouseUp = function (e) {
                var t = y._cinfo;
                y.lineDiv.removeEventListener("mouseup", y._mouseUp, !1), 5 < Math.abs(e.clientX - t.clientX) || 5 < Math.abs(e.clientY - t.clientY) || "function" == typeof y.handler && !1 === y.handler(t, y.cm) || he(t, y.cm);
            }, this._mouseDown = function (e) {
                var t = e.button, n = e.clientX, r = e.clientY, i = e.ctrlKey, o = e.altKey, a = e.shiftKey, l = y.cm;
                if ("PRE" !== e.target.tagName) {
                    var s, c, d, h, u = l.coordsChar({left: n, top: r}, "window"), f = l.getTokenAt(u), m = f.state,
                        p = " " + f.type + " ", g = null;
                    if (c = p.match(/\s(image|link|url)\s/)) {
                        g = c[1];
                        var v, b = /\shmd-barelink\s/.test(p);
                        if (m.linkText ? (s = w(l, u, function (e) {
                            return e.state.linkText || /(?:\s|^)link(?:\s|$)/.test(e.type);
                        }), g = "link") : s = w(l, u, g), /^(?:image|link)$/.test(g) && !b) {
                            var _ = w(l, {line: u.line, ch: s.to.ch + 1}, "url");
                            _ && (s.to = _.to);
                        }
                        if (")" === (d = l.getRange(s.from, s.to).trim()).slice(-1) && -1 !== (v = d.lastIndexOf("]("))) h = N(d.slice(v + 2, -1)).url; else if ((c = d.match(/[^\\]\]\s?\[([^\]]+)\]$/)) || (c = d.match(/^\[(.+)\]\s?\[\]$/)) || (c = d.match(/^\[(.+)\](?:\:\s*)?$/))) {
                            b && "^" === c[1].charAt(0) && (g = "footref");
                            var k = l.hmdReadLink(c[1], u.line);
                            h = k ? N(k.content).url : null;
                        } else ((c = d.match(/^\<(.+)\>$/)) || (c = d.match(/^\((.+)\)$/)) || (c = [null, d])) && (h = c[1]);
                        h = l.hmdResolveURL(h);
                    } else p.match(/\sformatting-task\s/) ? (g = "todo", (s = w(l, u, "formatting-task")).to.ch = l.getLine(u.line).length, d = l.getRange(s.from, s.to), h = null) : p.match(/\shashtag/) && (s = w(l, u, g = "hashtag"), d = l.getRange(s.from, s.to), h = null);
                    null !== g && (y._cinfo = {
                        type: g,
                        text: d,
                        url: h,
                        pos: u,
                        button: t,
                        clientX: n,
                        clientY: r,
                        ctrlKey: i,
                        altKey: o,
                        shiftKey: a
                    }, y.lineDiv.addEventListener("mouseup", y._mouseUp, !1));
                }
            }, this.lineDiv = e.display.lineDiv;
            var t = this.el = e.getWrapperElement();
            new s(function () {
                y.lineDiv.addEventListener("mousedown", y._mouseDown, !1), t.addEventListener("keydown", y._keyDown, !1);
            }, function () {
                y.lineDiv.removeEventListener("mousedown", y._mouseDown, !1), t.removeEventListener("keydown", y._keyDown, !1);
            }).bind(this, "enabled", !0);
        }, ge = x("Click", pe, fe),
        ve = Object.freeze({defaultClickHandler: he, defaultOption: fe, suggestedOption: me, Click: pe, getAddon: ge}),
        be = {enabled: !1, convertor: null}, _e = {enabled: !0};
    l.hmdPaste = _e, U.defineOption("hmdPaste", be, function (e, t) {
        t && "boolean" != typeof t ? "function" == typeof t && (t = {enabled: !0, convertor: t}) : t = {enabled: !!t};
        var n = Te(e);
        for (var r in be) n[r] = r in t ? t[r] : be[r];
    });
    var ke, ye, we = function (e) {
            var o = this;
            this.cm = e, this.pasteHandler = function (e, t) {
                var n = t.clipboardData || window.clipboardData, r = o.convertor;
                if (r && n && -1 != n.types.indexOf("text/html")) {
                    var i = r(n.getData("text/html"));
                    i && (e.operation(e.replaceSelection.bind(e, i)), t.preventDefault());
                }
            }, new s(function () {
                e.on("paste", o.pasteHandler);
            }, function () {
                e.off("paste", o.pasteHandler);
            }).bind(this, "enabled", !0);
        }, Te = x("Paste", we, be), xe = Object.freeze({defaultOption: be, suggestedOption: _e, Paste: we, getAddon: Te}),
        Le = "undefined" == typeof Uint8Array ? Array : Uint8Array;
    (ye = ke || (ke = {})).OK = "ok", ye.CURSOR_INSIDE = "ci", ye.HAS_MARKERS = "hm";
    var Me = {};

    function Ce(e, t, n, r) {
        var i = Me;
        if (e in i && !r) throw new Error("Folder " + e + " already registered");
        He[e] = !1, Se[e] = !!n, i[e] = t;
    }

    function Oe(t, n, r) {
        t.operation(function () {
            var e = n.find().from;
            e = {line: e.line, ch: e.ch + ~~r}, t.setCursor(e), t.focus(), n.clear();
        });
    }

    var He = {}, Se = {};
    l.hmdFold = Se, c.hmdFold = !1, U.defineOption("hmdFold", He, function (e, t) {
        t && "boolean" != typeof t || (t = t ? Se : He), "customFolders" in t && (console.error("[HyperMD][Fold] `customFolders` is removed. To use custom folders, `registerFolder` first."), delete t.customFolders);
        var n = De(e);
        for (var r in Me) n.setStatus(r, t[r]);
    });
    var Ee = function (t) {
        function e(e) {
            var u = this;
            t.call(this, e), this.cm = e, this._enabled = {}, this.folded = {}, this.startFold = h(this.startFoldImmediately.bind(this), 100), this._quickFoldHint = [], e.on("changes", function (e, t) {
                for (var n = [], r = 0, i = t; r < i.length; r += 1) for (var o = i[r], a = 0, l = e.findMarks(o.from, o.to); a < l.length; a += 1) {
                    var s = l[a];
                    s._hmd_fold_type && n.push(s);
                }
                for (var c = 0, d = n; c < d.length; c += 1) {
                    d[c].clear();
                }
                u.startFold();
            }), e.on("cursorActivity", function (c) {
                var d = {};

                function t(e) {
                    var t = e.line;
                    if (t in d) d[t].ch.push(e.ch); else {
                        for (var n = c.getLineHandle(e.line), r = n.markedSpans || [], i = [], o = 0; o < r.length; o++) {
                            var a = r[o].marker;
                            if ("_hmd_crange" in a) {
                                var l = a._hmd_crange[0].line < t ? 0 : a._hmd_crange[0].ch,
                                    s = a._hmd_crange[1].line > t ? n.text.length : a._hmd_crange[1].ch;
                                i.push([a, l, s]);
                            }
                        }
                        d[t] = {lineNo: t, ch: [e.ch], markers: i};
                    }
                }

                for (var e in c.listSelections().forEach(function (e) {
                    t(e.anchor), t(e.head);
                }), d) {
                    var n = d[e];
                    if (n.markers.length) for (var r = 0; r < n.ch.length; r++) for (var i = n.ch[r], o = 0; o < n.markers.length; o++) {
                        var a = n.markers[o], l = a[0], s = a[1], h = a[2];
                        s <= i && i <= h && (l.clear(), n.markers.splice(o, 1), o--);
                    }
                }
                u.startQuickFold();
            });
        }

        return t && (e.__proto__ = t), ((e.prototype = Object.create(t && t.prototype)).constructor = e).prototype.setStatus = function (e, t) {
            e in Me && !this._enabled[e] != !t && (this._enabled[e] = !!t, t ? this.startFold() : this.clear(e));
        }, e.prototype.requestRange = function (e, t, n, r) {
            r || (r = t), n || (n = e);
            var i = this.cm;
            if (0 !== i.findMarks(e, t).length) return ke.HAS_MARKERS;
            this._quickFoldHint.push(e.line), this._lastCRange = [n, r];
            for (var o = i.listSelections(), a = 0; a < o.length; a++) {
                var l = g(o[a]);
                if (k(this._lastCRange, l) || k([e, t], l)) return ke.CURSOR_INSIDE;
            }
            return this._quickFoldHint.push(n.line), ke.OK;
        }, e.prototype.startFoldImmediately = function (e, t) {
            var y = this, n = this.cm;
            e = e || n.firstLine(), t = (t || n.lastLine()) + 1, this._quickFoldHint = [], this.setPos(e, 0, !0), n.operation(function () {
                return n.eachLine(e, t, function (e) {
                    var t = e.lineNo();
                    if (!(t < y.lineNo)) {
                        t > y.lineNo && y.setPos(t, 0);
                        var n = new Le(e.text.length), r = e.markedSpans;
                        if (r) for (var i = 0; i < r.length; ++i) for (var o = r[i], a = null == o.from ? 0 : o.from, l = null == o.to ? n.length : o.to, s = a; s < l; s++) n[s] = 1;
                        for (var c = y.lineTokens; y.i_token < c.length;) {
                            for (var d, h = c[y.i_token], u = null, f = !0, m = h.start; m < h.end; m++) if (n[m]) {
                                f = !1;
                                break;
                            }
                            if (f) for (d in Me) if (y._enabled[d] && (u = Me[d](y, h))) break;
                            if (u) {
                                var p = u.find(), g = p.from, v = p.to;
                                if ((y.folded[d] || (y.folded[d] = [])).push(u), u._hmd_fold_type = d, u._hmd_crange = y._lastCRange, u.on("clear", function (e, t) {
                                    var n, r = y.folded[d];
                                    r && -1 !== (n = r.indexOf(u)) && r.splice(n, 1), y._quickFoldHint.push(e.line);
                                }), g.line > t || g.ch > h.start) {
                                    y.i_token++;
                                    for (var b = g.line === t ? g.ch : n.length, _ = v.line === t ? v.ch : n.length, k = b; k < _; k++) n[k] = 1;
                                } else {
                                    if (v.line !== t) return void y.setPos(v.line, v.ch);
                                    y.setPos(v.ch);
                                }
                            } else y.i_token++;
                        }
                    }
                });
            });
        }, e.prototype.startQuickFold = function () {
            var e = this._quickFoldHint;
            if (0 !== e.length) {
                for (var t = e[0], n = t, r = 0, i = e; r < i.length; r += 1) {
                    var o = i[r];
                    o < t && (t = o), n < o && (n = o);
                }
                this.startFold.stop(), this.startFoldImmediately(t, n);
            }
        }, e.prototype.clear = function (e) {
            this.startFold.stop();
            var t = this.folded[e];
            if (t && t.length) for (var n; n = t.pop();) n.clear();
        }, e.prototype.clearAll = function () {
            for (var e in this.startFold.stop(), this.folded) for (var t, n = this.folded[e]; t = n.pop();) t.clear();
        }, e;
    }(z), De = x("Fold", Ee), Ne = Object.freeze({
        get RequestRangeResult() {
            return ke;
        },
        folderRegistry: Me,
        registerFolder: Ce,
        breakMark: Oe,
        defaultOption: He,
        suggestedOption: Se,
        Fold: Ee,
        getAddon: De
    }), Ae = function (e, t) {
        var n = e.cm, r = /\bformatting-link-string\b/;
        if (/\bimage-marker\b/.test(t.type) && "!" === t.string) {
            var i = e.lineNo, o = e.findNext(r), a = e.findNext(r, o.i_token + 1), l = {line: i, ch: t.start},
                s = {line: i, ch: a.token.end}, c = e.requestRange(l, s, l, l);
            if (c === ke.OK) {
                var d, h, u = n.getRange({line: i, ch: o.token.start + 1}, {line: i, ch: a.token.start});
                if ("]" === a.token.string) {
                    var f = n.hmdReadLink(u, i);
                    if (!f) return null;
                    u = f.content;
                }
                d = N(u).url, d = n.hmdResolveURL(d), h = n.getRange({line: i, ch: l.ch + 2}, {
                    line: i,
                    ch: o.token.start - 1
                });
                var m = document.createElement("img"),
                    p = n.markText(l, s, {clearOnEnter: !0, collapsed: !0, replacedWith: m});
                return m.addEventListener("load", function () {
                    m.classList.remove("hmd-image-loading"), p.changed();
                }, !1), m.addEventListener("error", function () {
                    m.classList.remove("hmd-image-loading"), m.classList.add("hmd-image-error"), p.changed();
                }, !1), m.addEventListener("click", function () {
                    return Oe(n, p);
                }, !1), m.className = "hmd-image hmd-image-loading", m.src = d, m.title = h, p;
            }
        }
        return null;
    };
    Ce("image", Ae, !0);
    var Re = Object.freeze({ImageFolder: Ae}), Ie = function (e, t) {
        var n = e.cm;
        if ("[" !== t.string || !t.state.linkText || t.state.linkTitle || /\bimage\b/.test(t.type)) return null;
        var r = T(n), i = r.findSpanWithTypeAt({line: e.lineNo, ch: t.start}, "linkText");
        if (!i) return null;
        var o = r.findSpanWithTypeAt({line: e.lineNo, ch: i.end + 1}, "linkHref");
        if (!o) return null;
        var a = {line: e.lineNo, ch: o.begin}, l = {line: e.lineNo, ch: o.end}, s = {line: e.lineNo, ch: i.begin};
        if (e.requestRange(a, l, s, a) !== ke.OK) return null;
        var c = n.getRange(a, l), d = N(c.substr(1, c.length - 2)), h = d.url, u = d.title,
            f = document.createElement("span");
        f.setAttribute("class", "hmd-link-icon"), f.setAttribute("title", h + "\n" + u), f.setAttribute("data-url", h);
        var m = n.markText(a, l, {collapsed: !0, replacedWith: f});
        return f.addEventListener("click", function () {
            return Oe(n, m);
        }, !1), m;
    };
    Ce("link", Ie, !0);
    var je = Object.freeze({LinkFolder: Ie}), Fe = {};
    var Pe = function (e, t) {
        return 0 === t.start && t.type && -1 !== t.type.indexOf("HyperMD-codeblock-begin") && /[-\w]+\s*$/.test(t.string) ? Xe(e.cm).fold(e, t) : null;
    };
    Ce("code", Pe, !0);
    var qe = {}, $e = {};
    l.hmdFoldCode = $e, U.defineOption("hmdFoldCode", qe, function (e, t) {
        t && "boolean" != typeof t || (t = t ? $e : qe);
        var n = Xe(e);
        for (var r in Fe) n.setStatus(r, t[r]);
    });
    var ze = function (e) {
        this.cm = e, this._enabled = {}, this.folded = {};
    };
    ze.prototype.setStatus = function (e, t) {
        e in Fe && !this._enabled[e] != !t && (this._enabled[e] = !!t, t ? De(this.cm).startFold() : this.clear(e));
    }, ze.prototype.clear = function (e) {
        var t = this.folded[e];
        if (t && t.length) for (var n; n = t.pop();) n.marker.clear();
    }, ze.prototype.fold = function (e, t) {
        var n = this;
        if (0 !== t.start || !t.type || -1 === t.type.indexOf("HyperMD-codeblock-begin")) return null;
        var r, i, o = /([-\w]+)\s*$/.exec(t.string), a = o && o[1].toLowerCase();
        if (!a) return null;
        var l = this.cm, s = Fe, c = this._enabled;
        for (var d in s) {
            var h = s[d];
            if (c[d] && h.pattern(a)) {
                i = d, r = h.renderer;
                break;
            }
        }
        if (!i) return null;
        var u = {line: e.lineNo, ch: 0}, f = null, m = l.lastLine(), p = e.lineNo + 1;
        do {
            var g = l.getTokenAt({line: p, ch: 1});
            if (g && g.type && -1 !== g.type.indexOf("HyperMD-codeblock-end")) {
                f = {line: p, ch: g.end};
                break;
            }
        } while (++p < m);
        if (!f) return null;
        if (e.requestRange(u, f) !== ke.OK) return null;
        var v = l.getRange({line: u.line + 1, ch: 0}, {line: f.line, ch: 0}),
            b = {editor: l, lang: a, marker: null, lineWidget: null, el: null, break: Be, changed: Be},
            _ = b.el = r(v, b);
        if (!_) return b.marker.clear(), null;
        var k = document.createElement("div");
        k.className = Ke + i, k.style.minHeight = "1em", k.appendChild(_);
        var y = b.lineWidget = l.addLineWidget(f.line, k, {
            above: !1,
            coverGutter: !1,
            noHScroll: !1,
            showIfHidden: !1
        }), w = document.createElement("span");
        w.className = Ue + i, w.textContent = "<CODE>";
        var T = b.marker = l.markText(u, f, {replacedWith: w});
        return k.addEventListener("mouseenter", function () {
            return w.className = We + i;
        }, !1), k.addEventListener("mouseleave", function () {
            return w.className = Ue + i;
        }, !1), b.changed = function () {
            y.changed();
        }, b.break = function () {
            Oe(l, T);
        }, w.addEventListener("click", b.break, !1), T.on("clear", function () {
            var e, t = n.folded[i];
            t && -1 !== (e = t.indexOf(b)) && t.splice(e, 1), "function" == typeof b.onRemove && b.onRemove(b), y.clear();
        }), i in this.folded ? this.folded[i].push(b) : this.folded[i] = [b], T;
    };
    var Ke = "hmd-fold-code-content hmd-fold-code-", Ue = "hmd-fold-code-stub hmd-fold-code-",
        We = "hmd-fold-code-stub highlight hmd-fold-code-", Be = function () {
        }, Xe = x("FoldCode", ze, qe), Ye = Object.freeze({
            rendererRegistry: Fe, registerRenderer: function (t, e) {
                if (t && t.name && t.renderer) {
                    var n = t.name, r = t.pattern, i = Fe;
                    if (n in i && !e) throw new Error("CodeRenderer " + n + " already exists");
                    if ("string" == typeof r) {
                        var o = r.toLowerCase();
                        r = function (e) {
                            return e.toLowerCase() === o;
                        };
                    } else r instanceof RegExp && (r = function (e) {
                        return t.pattern.test(e);
                    });
                    var a = {name: n, suggested: !!t.suggested, pattern: r, renderer: t.renderer};
                    i[n] = a, qe[n] = !1, $e[n] = a.suggested;
                }
            }, CodeFolder: Pe, defaultOption: qe, suggestedOption: $e, FoldCode: ze, getAddon: Xe
        }), Ge = function (e, t) {
            var n = /formatting-math-begin\b/;
            if (!n.test(t.type)) return null;
            var r = e.cm, i = e.lineNo, o = /math-2\b/.test(t.type), a = o ? 2 : 1;
            if (2 == a && 1 == t.string.length) {
                0;
                var l = e.lineTokens[e.i_token + 1];
                if (!l || !n.test(l.type)) return null;
            }
            var s, c = e.findNext(/formatting-math-end\b/, o), d = {line: i, ch: t.start}, h = !1;
            if (c) s = {line: c.lineNo, ch: c.token.start + a}; else {
                if (!o) return null;
                var u = r.lastLine();
                s = {line: u, ch: r.getLine(u).length}, h = !0;
            }
            var f = {line: d.line, ch: d.ch + a}, m = {line: s.line, ch: s.ch - (h ? 0 : a)}, p = r.getRange(f, m).trim(),
                g = tt(r), v = e.requestRange(d, s);
            if (v !== ke.OK) return v === ke.CURSOR_INSIDE && (g.editingExpr = p), null;
            var b = Ve(r, d, s, p, a, 1 < a && 0 == d.ch && (h || s.ch >= r.getLine(s.line).length));
            return g.editingExpr = null, b;
        };

    function Ve(e, t, n, r, i, o) {
        var a = document.createElement("span");
        a.setAttribute("class", "hmd-fold-math math-" + (o ? 2 : 1)), a.setAttribute("title", r);
        var l = document.createElement("span");
        l.setAttribute("class", "hmd-fold-math-placeholder"), l.textContent = r, a.appendChild(l);
        var s = e.markText(t, n, {replacedWith: a});
        a.addEventListener("click", function () {
            return Oe(e, s, i);
        }, !1);
        var c = tt(e).createRenderer(a, o ? "display" : "");
        return c.onChanged = function () {
            l && (a.removeChild(l), l = null), s.changed();
        }, s.on("clear", function () {
            c.clear();
        }), s.mathRenderer = c, d(function () {
            return !!c.isReady() && (c.startRender(r), !0);
        }, 5, function () {
            s.clear();
        }), s;
    }

    Ce("math", Ge, !0);
    var Ze = function (e, t) {
        var n = this;
        this.container = e;
        var r = document.createElement("img");
        r.setAttribute("class", "hmd-math-dumb"), r.addEventListener("load", function () {
            n.onChanged && n.onChanged(n.last_expr);
        }, !1), this.img = r, e.appendChild(r);
    };
    Ze.prototype.startRender = function (e) {
        this.last_expr = e, this.img.src = "https://latex.codecogs.com/gif.latex?" + encodeURIComponent(e);
    }, Ze.prototype.clear = function () {
        this.container.removeChild(this.img);
    }, Ze.prototype.isReady = function () {
        return !0;
    };
    var Qe = {renderer: Ze, onPreview: null, onPreviewEnd: null}, Je = {};
    l.hmdFoldMath = Je, U.defineOption("hmdFoldMath", Qe, function (e, t) {
        t ? "function" == typeof t && (t = {renderer: t}) : t = {};
        var n = tt(e);
        for (var r in Qe) n[r] = r in t ? t[r] : Qe[r];
    });
    var et = function (e) {
        var t = this;
        this.cm = e, new s(function (e) {
            t.onPreview && t.onPreview(e);
        }, function () {
            t.onPreviewEnd && t.onPreviewEnd();
        }, null).bind(this, "editingExpr");
    };
    et.prototype.createRenderer = function (e, t) {
        return new (this.renderer || Ze)(e, t);
    };
    var tt = x("FoldMath", et, Qe), nt = Object.freeze({
        MathFolder: Ge,
        insertMathMark: Ve,
        DumbRenderer: Ze,
        defaultOption: Qe,
        suggestedOption: Je,
        FoldMath: et,
        getAddon: tt
    }), rt = {}, it = function (e) {
        return e in rt;
    }, ot = function (e) {
        var t = document.createElement("span");
        return t.textContent = rt[e], t.title = e, t;
    }, at = function (e, t) {
        if (!t.type || !/ formatting-emoji/.test(t.type)) return null;
        var n = e.cm, r = {line: e.lineNo, ch: t.start}, i = {line: e.lineNo, ch: t.end}, o = t.string, a = dt(n);
        return a.isEmoji(o) ? e.requestRange(r, i) !== ke.OK ? null : a.foldEmoji(o, r, i) : null;
    };
    Ce("emoji", at, !0);
    var lt = {myEmoji: {}, emojiRenderer: ot, emojiChecker: it}, st = {};
    l.hmdFoldEmoji = st, U.defineOption("hmdFoldEmoji", lt, function (e, t) {
        t || (t = {});
        var n = dt(e);
        for (var r in lt) n[r] = r in t ? t[r] : lt[r];
    });
    var ct = function (e) {
        this.cm = e;
    };
    ct.prototype.isEmoji = function (e) {
        return e in this.myEmoji || this.emojiChecker(e);
    }, ct.prototype.foldEmoji = function (e, t, n) {
        var r = this.cm, i = e in this.myEmoji && this.myEmoji[e](e) || this.emojiRenderer(e);
        if (!i || !i.tagName) return null;
        -1 === i.className.indexOf("hmd-emoji") && (i.className += " hmd-emoji");
        var o = r.markText(t, n, {replacedWith: i});
        return i.addEventListener("click", Oe.bind(this, r, o, 1), !1), "img" === i.tagName.toLowerCase() && (i.addEventListener("load", function () {
            return o.changed();
        }, !1), i.addEventListener("dragstart", function (e) {
            return e.preventDefault();
        }, !1)), o;
    };
    var dt = x("FoldEmoji", ct, lt);
    !function (e) {
        for (var t, n = ["smile:😄;laughing:😆;blush:😊;smiley:😃;relaxed:☺️;smirk:😏;heart_eyes:😍;kissing_heart:😘;kissing_closed_eyes:😚;flushed:😳;relieved:😌;satisfied:😆;grin:😁;wink:😉;stuck_out_tongue_winking_eye:😜;stuck_out_tongue_closed_eyes:😝;grinning:😀;kissing:😗;kissing_smiling_eyes:😙;stuck_out_tongue:😛;sleeping:😴;worried:😟;frowning:😦;anguished:😧;open_mouth:😮;grimacing:😬;confused:😕;hushed:😯;expressionless:😑;unamused:😒;sweat_smile:😅;sweat:😓;disappointed_relieved:😥;weary:😩;pensive:😔;disappointed:😞;confounded:😖;fearful:😨;cold_sweat:😰;persevere:😣;cry:😢;sob:😭;joy:😂;astonished:😲;scream:😱;tired_face:😫;angry:😠;rage:😡;triumph:😤;sleepy:😪;yum:😋;mask:😷;sunglasses:😎;dizzy_face:😵;imp:👿;smiling_imp:😈;neutral_face:😐;no_mouth:😶;innocent:😇;alien:👽;yellow_heart:💛;blue_heart:💙;purple_heart:💜;heart:❤️;green_heart:💚;broken_heart:💔;heartbeat:💓;heartpulse:💗;two_hearts:💕;revolving_hearts:💞;cupid:💘;sparkling_heart:💖;sparkles:✨;star:⭐️;star2:🌟;dizzy:💫;boom:💥;collision:💥;anger:💢;exclamation:❗️;question:❓;grey_exclamation:❕;grey_question:❔;zzz:💤;dash:💨;sweat_drops:💦;notes:🎶;musical_note:🎵;fire:🔥;hankey:💩;poop:💩;shit:💩;", "+1:👍;thumbsup:👍;-1:👎;thumbsdown:👎;ok_hand:👌;punch:👊;facepunch:👊;fist:✊;v:✌️;wave:👋;hand:✋;raised_hand:✋;open_hands:👐;point_up:☝️;point_down:👇;point_left:👈;point_right:👉;raised_hands:🙌;pray:🙏;point_up_2:👆;clap:👏;muscle:💪;metal:🤘;fu:🖕;walking:🚶;runner:🏃;running:🏃;couple:👫;family:👪;two_men_holding_hands:👬;two_women_holding_hands:👭;dancer:💃;dancers:👯;ok_woman:🙆;no_good:🙅;information_desk_person:💁;raising_hand:🙋;bride_with_veil:👰;person_with_pouting_face:🙎;person_frowning:🙍;bow:🙇;couplekiss::couplekiss:;couple_with_heart:💑;massage:💆;haircut:💇;nail_care:💅;boy:👦;girl:👧;woman:👩;man:👨;baby:👶;older_woman:👵;older_man:👴;person_with_blond_hair:👱;man_with_gua_pi_mao:👲;man_with_turban:👳;construction_worker:👷;cop:👮;angel:👼;princess:👸;smiley_cat:😺;smile_cat:😸;heart_eyes_cat:😻;kissing_cat:😽;smirk_cat:😼;scream_cat:🙀;crying_cat_face:😿;joy_cat:😹;pouting_cat:😾;japanese_ogre:👹;japanese_goblin:👺;see_no_evil:🙈;hear_no_evil:🙉;speak_no_evil:🙊;guardsman:💂;skull:💀;feet:🐾;lips:👄;kiss:💋;droplet:💧;ear:👂;eyes:👀;nose:👃;tongue:👅;love_letter:💌;bust_in_silhouette:👤;busts_in_silhouette:👥;speech_balloon:💬;", "thought_balloon:💭;sunny:☀️;umbrella:☔️;cloud:☁️;snowflake:❄️;snowman:⛄️;zap:⚡️;cyclone:🌀;foggy:🌁;ocean:🌊;cat:🐱;dog:🐶;mouse:🐭;hamster:🐹;rabbit:🐰;wolf:🐺;frog:🐸;tiger:🐯;koala:🐨;bear:🐻;pig:🐷;pig_nose:🐽;cow:🐮;boar:🐗;monkey_face:🐵;monkey:🐒;horse:🐴;racehorse:🐎;camel:🐫;sheep:🐑;elephant:🐘;panda_face:🐼;snake:🐍;bird:🐦;baby_chick:🐤;hatched_chick:🐥;hatching_chick:🐣;chicken:🐔;penguin:🐧;turtle:🐢;bug:🐛;honeybee:🐝;ant:🐜;beetle:🐞;snail:🐌;octopus:🐙;tropical_fish:🐠;fish:🐟;whale:🐳;whale2:🐋;dolphin:🐬;cow2:🐄;ram:🐏;rat:🐀;water_buffalo:🐃;tiger2:🐅;rabbit2:🐇;dragon:🐉;goat:🐐;rooster:🐓;dog2:🐕;pig2:🐖;mouse2:🐁;ox:🐂;dragon_face:🐲;blowfish:🐡;crocodile:🐊;dromedary_camel:🐪;leopard:🐆;cat2:🐈;poodle:🐩;paw_prints:🐾;bouquet:💐;cherry_blossom:🌸;tulip:🌷;four_leaf_clover:🍀;rose:🌹;sunflower:🌻;hibiscus:🌺;maple_leaf:🍁;leaves:🍃;fallen_leaf:🍂;herb:🌿;mushroom:🍄;cactus:🌵;palm_tree:🌴;evergreen_tree:🌲;deciduous_tree:🌳;chestnut:🌰;seedling:🌱;blossom:🌼;ear_of_rice:🌾;shell:🐚;globe_with_meridians:🌐;sun_with_face:🌞;full_moon_with_face:🌝;new_moon_with_face:🌚;new_moon:🌑;waxing_crescent_moon:🌒;first_quarter_moon:🌓;waxing_gibbous_moon:🌔;full_moon:🌕;waning_gibbous_moon:🌖;last_quarter_moon:🌗;waning_crescent_moon:🌘;last_quarter_moon_with_face:🌜;", "first_quarter_moon_with_face:🌛;moon:🌔;earth_africa:🌍;earth_americas:🌎;earth_asia:🌏;volcano:🌋;milky_way:🌌;partly_sunny:⛅️;bamboo:🎍;gift_heart:💝;dolls:🎎;school_satchel:🎒;mortar_board:🎓;flags:🎏;fireworks:🎆;sparkler:🎇;wind_chime:🎐;rice_scene:🎑;jack_o_lantern:🎃;ghost:👻;santa:🎅;christmas_tree:🎄;gift:🎁;bell:🔔;no_bell:🔕;tanabata_tree:🎋;tada:🎉;confetti_ball:🎊;balloon:🎈;crystal_ball:🔮;cd:💿;dvd:📀;floppy_disk:💾;camera:📷;video_camera:📹;movie_camera:🎥;computer:💻;tv:📺;iphone:📱;phone:☎️;telephone:☎️;telephone_receiver:📞;pager:📟;fax:📠;minidisc:💽;vhs:📼;sound:🔉;speaker:🔈;mute:🔇;loudspeaker:📢;mega:📣;hourglass:⌛️;hourglass_flowing_sand:⏳;alarm_clock:⏰;watch:⌚️;radio:📻;satellite:📡;loop:➿;mag:🔍;mag_right:🔎;unlock:🔓;lock:🔒;lock_with_ink_pen:🔏;closed_lock_with_key:🔐;key:🔑;bulb:💡;flashlight:🔦;high_brightness:☘️;low_brightness:🔅;electric_plug:🔌;battery:🔋;calling:📲;email:✉️;mailbox:📫;postbox:📮;bath:🛀;bathtub:🛁;shower:🚿;toilet:🚽;wrench:🔧;nut_and_bolt:🔩;hammer:🔨;seat:💺;moneybag:💰;yen:💴;dollar:💵;pound:💷;euro:💶;", "credit_card:💳;money_with_wings:💸;e-mail:📧;inbox_tray:📥;outbox_tray:📤;envelope:✉️;incoming_envelope:📨;postal_horn:📯;mailbox_closed:📪;mailbox_with_mail:📬;mailbox_with_no_mail:📭;door:🚪;smoking:🚬;bomb:💣;gun:🔫;hocho:🔪;pill:💊;syringe:💉;page_facing_up:📄;page_with_curl:📃;bookmark_tabs:📑;bar_chart:📊;chart_with_upwards_trend:📈;chart_with_downwards_trend:📉;scroll:📜;clipboard:📋;calendar:📆;date:📅;card_index:📇;file_folder:📁;open_file_folder:📂;scissors:✂️;pushpin:📌;paperclip:📎;black_nib:✒️;pencil2:✏️;straight_ruler:📏;triangular_ruler:📐;closed_book:📕;green_book:📗;blue_book:📘;orange_book:📙;notebook:📓;notebook_with_decorative_cover:📔;ledger:📒;books:📚;bookmark:🔖;name_badge:📛;microscope:🏗;telescope:🔭;newspaper:📰;football:🏈;basketball:🏀;soccer:⚽️;baseball:⚾️;tennis:🎾;8ball:🎱;", "rugby_football:🏉;bowling:🎳;golf:⛳️;mountain_bicyclist:🚵;bicyclist:🚴;horse_racing:🏇;snowboarder:🏂;swimmer:🏊;surfer:🏄;ski:🎿;spades:♠️;hearts:♥️;clubs:♣️;diamonds:♦️;gem:💎;ring:💍;trophy:🏆;musical_score:🎼;musical_keyboard:🎹;violin:🎻;space_invader:👾;video_game:🎮;black_joker:🃏;flower_playing_cards:🎴;game_die:🎲;dart:🎯;mahjong:🀄️;clapper:🎬;memo:📝;pencil:📝;book:📖;art:🎨;microphone:🎤;headphones:🎧;trumpet:🎺;saxophone:🎷;guitar:🎸;shoe:👞;sandal:👡;high_heel:👠;lipstick:💄;boot:👢;shirt:👕;tshirt:👕;necktie:👔;womans_clothes:👚;dress:👗;running_shirt_with_sash:🎽;jeans:👖;kimono:👘;bikini:👙;ribbon:🎀;tophat:🎩;crown:👑;womans_hat:👒;mans_shoe:👞;closed_umbrella:🌂;briefcase:💼;handbag:👜;pouch:👝;purse:👛;eyeglasses:👓;fishing_pole_and_fish:🎣;coffee:☕️;tea:🍵;sake:🍶;baby_bottle:🍼;beer:🍺;beers:🍻;cocktail:🍸;tropical_drink:🍹;wine_glass:🍷;fork_and_knife:🍴;pizza:🍕;hamburger:🍔;fries:🍟;poultry_leg:🍗;meat_on_bone:🍖;spaghetti:🍝;curry:🍛;fried_shrimp:🍤;bento:🍱;sushi:🍣;fish_cake:🍥;rice_ball:🍙;rice_cracker:🍘;rice:🍚;", "ramen:🍜;stew:🍲;oden:🍢;dango:🍡;egg:🥚;bread:🍞;doughnut:🍩;custard:🍮;icecream:🍦;ice_cream:🍨;shaved_ice:🍧;birthday:🎂;cake:🍰;cookie:🍪;chocolate_bar:🍫;candy:🍬;lollipop:🍭;honey_pot:🍯;apple:🍎;green_apple:🍏;tangerine:🍊;lemon:🍋;cherries:🍒;grapes:🍇;watermelon:🍉;strawberry:🍓;peach:🍑;melon:🍈;banana:🍌;pear:🍐;pineapple:🍍;sweet_potato:🍠;eggplant:🍆;tomato:🍅;corn:🌽;house:🏠;house_with_garden:🏡;school:🏫;office:🏢;post_office:🏣;hospital:🏥;bank:🏦;convenience_store:🏪;love_hotel:🏩;hotel:🏨;wedding:💒;church:⛪️;department_store:🏬;european_post_office:🏤;city_sunrise:🌇;city_sunset:🌆;japanese_castle:🏯;european_castle:🏰;tent:⛺️;factory:🏭;tokyo_tower:🗼;japan:🗾;mount_fuji:🗻;sunrise_over_mountains:🌄;sunrise:🌅;stars:🌠;statue_of_liberty:🗽;bridge_at_night:🌉;carousel_horse:🎠;rainbow:🌈;ferris_wheel:🎡;fountain:⛲️;roller_coaster:🎢;ship:🚢;speedboat:🚤;boat:⛵️;sailboat:⛵️;rowboat:🚣;anchor:⚓️;rocket:🚀;airplane:✈️;helicopter:🚁;steam_locomotive:🚂;tram:🚊;mountain_railway:🚞;bike:🚲;aerial_tramway:🚡;suspension_railway:🚟;", "mountain_cableway:🚠;tractor:🚜;blue_car:🚙;oncoming_automobile:🚘;car:🚗;red_car:🚗;taxi:🚕;oncoming_taxi:🚖;articulated_lorry:🚛;bus:🚌;oncoming_bus:🚍;rotating_light:🚨;police_car:🚓;oncoming_police_car:🚔;fire_engine:🚒;ambulance:🚑;minibus:🚐;truck:🚚;train:🚋;station:🚉;train2:🚆;bullettrain_front:🚅;bullettrain_side:🚄;light_rail:🚈;monorail:🚝;railway_car:🚃;trolleybus:🚎;ticket:🎫;fuelpump:⛽️;vertical_traffic_light:🚦;traffic_light:🚥;warning:⚠️;construction:🚧;beginner:🔰;atm:🏧;slot_machine:🎰;busstop:🚏;barber:💈;hotsprings:♨️;checkered_flag:🏁;crossed_flags:🎌;izakaya_lantern:🏮;moyai:🗿;circus_tent:🎪;performing_arts:🎭;round_pushpin:📍;triangular_flag_on_post:🚩;jp:🇯🇵;kr:🇰🇷;cn:🇨🇳;us:🇺🇸;fr:🇫🇷;es:🇪🇸;it:🇮🇹;ru:🇷🇺;gb:🇬🇧;uk:🇬🇧;de:🇩🇪;one:1️⃣;two:2️⃣;three:3️⃣;four:4️⃣;five:5️⃣;six:6️⃣;seven:7️⃣;eight:8️⃣;nine:9️⃣;keycap_ten:🔟;", "1234:🔢;zero:0️⃣;hash:#️⃣;symbols:🔣;arrow_backward:◀️;arrow_down:⬇️;arrow_forward:▶️;arrow_left:⬅️;capital_abcd:🔠;abcd:🔡;abc:🔤;arrow_lower_left:↙️;arrow_lower_right:↘️;arrow_right:➡️;arrow_up:⬆️;arrow_upper_left:↖️;arrow_upper_right:↗️;arrow_double_down:⏬;arrow_double_up:⏫;arrow_down_small:🔽;arrow_heading_down:⤵️;arrow_heading_up:⤴️;leftwards_arrow_with_hook:↩️;arrow_right_hook:↪️;left_right_arrow:↔️;arrow_up_down:↕️;arrow_up_small:🔼;arrows_clockwise:🔃;arrows_counterclockwise:🔄;rewind:⏪;fast_forward:⏩;information_source:ℹ️;ok:🆗;twisted_rightwards_arrows:🔀;repeat:🔁;repeat_one:🔂;new:🆕;top:🔝;up:🆙;cool:🆒;free:🆓;ng:🆖;cinema:🎦;koko:🈁;signal_strength:📶;u5272:🈹;u5408:🈴;u55b6:🈺;u6307:🈯️;u6708:🈷️;u6709:🈶;u6e80:🈵;u7121:🈚️;u7533:🈸;u7a7a:🈳;u7981:🈲;sa:🈂️;restroom:🚻;mens:🚹;womens:🚺;baby_symbol:🚼;no_smoking:🚭;", "parking:🅿️;wheelchair:♿️;metro:🚇;baggage_claim:🛄;accept:🉑;wc:🚾;potable_water:🚰;put_litter_in_its_place:🚮;secret:㊙️;congratulations:㊗️;m:Ⓜ️;passport_control:🛂;left_luggage:🛅;customs:🛃;ideograph_advantage:🉐;cl:🆑;sos:🆘;id:🆔;no_entry_sign:🚫;underage:🔞;no_mobile_phones:📵;do_not_litter:🚯;non-potable_water:🚱;no_bicycles:🚳;no_pedestrians:🚷;children_crossing:🚸;no_entry:⛔️;eight_spoked_asterisk:✳️;eight_pointed_black_star:✴️;heart_decoration:💟;vs:🆚;vibration_mode:📳;mobile_phone_off:📴;chart:💹;currency_exchange:💱;aries:♈️;taurus:♉️;gemini:♊️;cancer:♋️;leo:♌️;virgo:♍️;libra:♎️;scorpius:♏️;", "sagittarius:♐️;capricorn:♑️;aquarius:♒️;pisces:♓️;ophiuchus:⛎;six_pointed_star:🔯;negative_squared_cross_mark:❎;a:🅰️;b:🅱️;ab:🆎;o2:🅾️;diamond_shape_with_a_dot_inside:💠;recycle:♻️;end:🔚;on:🔛;soon:🔜;clock1:🕐;clock130:🕜;clock10:🕙;clock1030:🕥;clock11:🕚;clock1130:🕦;clock12:🕛;clock1230:🕧;clock2:🕑;clock230:🕝;clock3:🕒;clock330:🕞;clock4:🕓;clock430:🕟;clock5:🕔;clock530:🕠;clock6:🕕;clock630:🕡;clock7:🕖;clock730:🕢;clock8:🕗;clock830:🕣;clock9:🕘;clock930:🕤;heavy_dollar_sign:💲;copyright:©️;registered:®️;tm:™️;x:❌;heavy_exclamation_mark:❗️;bangbang:‼️;interrobang:⁉️;o:⭕️;heavy_multiplication_x:✖️;", "heavy_plus_sign:➕;heavy_minus_sign:➖;heavy_division_sign:➗;white_flower:💮;100:💯;heavy_check_mark:✔️;ballot_box_with_check:☑️;radio_button:🔘;link:🔗;curly_loop:➰;wavy_dash:〰️;part_alternation_mark:〽️;trident:🔱;black_square::black_square:;white_square::white_square:;white_check_mark:✅;black_square_button:🔲;white_square_button:🔳;black_circle:⚫️;white_circle:⚪️;red_circle:🔴;large_blue_circle:🔵;large_blue_diamond:🔷;large_orange_diamond:🔶;small_blue_diamond:🔹;small_orange_diamond:🔸;small_red_triangle:🔺;small_red_triangle_down:🔻"], r = /([-\w]+:)([^;]+);/g, i = 0; i < n.length; i++) for (r.lastIndex = 0; t = r.exec(n[i]);) e[":" + t[1]] = t[2];
    }(rt);
    var ht = Object.freeze({
        defaultDict: rt,
        defaultChecker: it,
        defaultRenderer: ot,
        EmojiFolder: at,
        defaultOption: lt,
        suggestedOption: st,
        FoldEmoji: ct,
        getAddon: dt
    }), ut = function (e) {
        return !/^<(?:br)/i.test(e) && (!/<(?:script|style|link|meta)/i.test(e) && (!/\son\w+\s*=/i.test(e) && !/src\s*=\s*["']?javascript:/i.test(e)));
    }, ft = function (e, t, a) {
        var n = /^<(\w+)\s*/.exec(e);
        if (!n) return null;
        for (var r, i = n[1], o = document.createElement(i), l = /([\w\:\-]+)(?:\s*=\s*((['"]).*?\3|\S+))?\s*/g, s = l.lastIndex = n[0].length; (r = l.exec(e)) && !(r.index > s);) {
            var c = r[1], d = r[2];
            d && /^['"]/.test(d) && (d = d.slice(1, -1)), o.setAttribute(c, d), s = l.lastIndex;
        }
        if ("innerHTML" in o) {
            var h = e.indexOf(">", s) + 1, u = e.length;
            (r = new RegExp("</" + i + "\\s*>\\s*$", "i").exec(e)) && (u = r.index);
            var f = e.slice(h, u);
            f && (o.innerHTML = f), m([o], function (e) {
                var t = e.tagName.toLowerCase();
                "a" === t && (e.getAttribute("target") || e.setAttribute("target", "_blank"));
                var n = {a: ["href"], img: ["src"], iframe: ["src"]}[t];
                if (n) for (var r = 0; r < n.length; r++) {
                    var i = n[r], o = e.getAttribute(i);
                    o && e.setAttribute(i, a.hmdResolveURL(o));
                }
            });
        }
        return o;
    }, mt = "hmd-fold-html-stub", pt = function (e, t) {
        if (!t.type || !/ hmd-html-begin/.test(t.type)) return null;
        var n = e.findNext(/ hmd-html-\w+/, !0);
        if (!n || !/ hmd-html-end/.test(n.token.type) || / hmd-html-unclosed/.test(n.token.type)) return null;
        var r = e.cm, i = {line: e.lineNo, ch: t.start}, o = {line: n.lineNo, ch: n.token.end},
            a = 0 != i.ch || o.ch < r.getLine(o.line).length, l = _t(r), s = r.getRange(i, o);
        return l.checker(s, i, r) ? e.requestRange(i, o) !== ke.OK ? null : l.renderAndInsert(s, i, o, a) : null;
    };
    Ce("html", pt, !1);
    var gt = {
        checker: ut,
        renderer: ft,
        stubText: "<HTML>",
        isolatedTagName: /^(?:div|pre|form|table|iframe|ul|ol|input|textarea|p|summary|a)$/i
    }, vt = {};
    l.hmdFoldHTML = vt, U.defineOption("hmdFoldHTML", gt, function (e, t) {
        t ? "function" == typeof t ? t = {checker: t} : "object" != typeof t && (console.warn("[HyperMD][FoldHTML] incorrect option value type"), t = {}) : t = {};
        var n = _t(e);
        for (var r in gt) n[r] = r in t ? t[r] : gt[r];
        !n.isolatedTagName || n.isolatedTagName instanceof RegExp || (console.error("[HyperMD][FoldHTML] option isolatedTagName only accepts RegExp"), n.isolatedTagName = gt.isolatedTagName);
    });
    var bt = function (e) {
        this.cm = e;
    };
    bt.prototype.renderAndInsert = function (e, t, n, r) {
        var i, o, a = this.cm, l = this.makeStub(), s = this.renderer(e, t, a), c = function () {
            return Oe(a, o);
        };
        if (!s) return null;
        if (l.addEventListener("click", c, !1), s.tagName.match(this.isolatedTagName || /^$/) || s.addEventListener("click", c, !1), r) {
            var d = document.createElement("span");
            d.setAttribute("class", "hmd-fold-html"), d.setAttribute("style", "display: inline-block"), d.appendChild(l), d.appendChild(s), i = d, (h = p(s, function (e, t) {
                var n = getComputedStyle(s), r = function (e) {
                    return n.getPropertyValue(e);
                }, i = e < 10 || t < 10 || !/^relative|static$/i.test(r("position")) || !/^none$/i.test(r("float"));
                l.className = i ? mt : "hmd-fold-html-stub omittable", o.changed();
            })).check(), setTimeout(function () {
                o.on("clear", function () {
                    h.stop();
                });
            }, 0);
        } else {
            i = l;
            var h, u = a.addLineWidget(n.line, s, {above: !1, coverGutter: !1, noHScroll: !1, showIfHidden: !1}),
                f = function () {
                    return l.className = "hmd-fold-html-stub highlight";
                }, m = function () {
                    return l.className = mt;
                };
            s.addEventListener("mouseenter", f, !1), s.addEventListener("mouseleave", m, !1), (h = p(s, function () {
                return u.changed();
            })).check(), setTimeout(function () {
                o.on("clear", function () {
                    h.stop(), u.clear(), s.removeEventListener("mouseenter", f, !1), s.removeEventListener("mouseleave", m, !1);
                });
            }, 0);
        }
        return o = a.markText(t, n, {replacedWith: i});
    }, bt.prototype.makeStub = function () {
        var e = document.createElement("span");
        return e.setAttribute("class", mt), e.textContent = this.stubText || "<HTML>", e;
    };
    var _t = x("FoldHTML", bt, gt), kt = Object.freeze({
        defaultChecker: ut,
        defaultRenderer: ft,
        HTMLFolder: pt,
        defaultOption: gt,
        suggestedOption: vt,
        FoldHTML: bt,
        getAddon: _t
    }), yt = {enabled: !1}, wt = {enabled: !0};
    l.hmdTableAlign = wt, c.hmdTableAlign = !1, U.defineOption("hmdTableAlign", yt, function (e, t) {
        var n = !!t;
        n && "boolean" != typeof t || (t = {enabled: n});
        var r = xt(e);
        for (var i in yt) r[i] = i in t ? t[i] : yt[i];
    });
    var Tt = function (e) {
        var p = this;
        this.cm = e, this.styleEl = document.createElement("style"), this.updateStyle = h(function () {
            if (p.enabled) {
                var e = p.cm, t = p.measure(), n = p.makeCSS(t);
                n !== p._lastCSS && (p.styleEl.textContent = p._lastCSS = n, e.refresh());
            }
        }, 100), this._procLine = function (e, t, n) {
            if (n.querySelector(".cm-hmd-table-sep")) {
                for (var r = n.firstElementChild, i = Array.prototype.slice.call(r.childNodes, 0), o = e.getStateAfter(t.lineNo()), a = o.hmdTableColumns, l = o.hmdTableID, s = 2 === o.hmdTable ? -1 : 0, c = p.makeColumn(s, a[s] || "dummy", l), d = c.firstElementChild, h = 0, u = i; h < u.length; h += 1) {
                    var f = u[h], m = f.nodeType === Node.ELEMENT_NODE && f.className || "";
                    /cm-hmd-table-sep/.test(m) ? (s++, c.appendChild(d), r.appendChild(c), r.appendChild(f), d = (c = p.makeColumn(s, a[s] || "dummy", l)).firstElementChild) : d.appendChild(f);
                }
                c.appendChild(d), r.appendChild(c);
            }
        }, new s(function () {
            e.on("renderLine", p._procLine), e.on("update", p.updateStyle), e.refresh(), document.head.appendChild(p.styleEl);
        }, function () {
            e.off("renderLine", p._procLine), e.off("update", p.updateStyle), document.head.removeChild(p.styleEl);
        }).bind(this, "enabled", !0);
    };
    Tt.prototype.makeColumn = function (e, t, n) {
        var r = document.createElement("span");
        r.className = "hmd-table-column hmd-table-column-" + e + " hmd-table-column-" + t, r.setAttribute("data-column", "" + e), r.setAttribute("data-table-id", n);
        var i = document.createElement("span");
        return i.className = "hmd-table-column-content", i.setAttribute("data-column", "" + e), r.appendChild(i), r;
    }, Tt.prototype.measure = function () {
        for (var e = this.cm.display.lineDiv.querySelectorAll(".hmd-table-column-content"), t = {}, n = 0; n < e.length; n++) {
            var r = e[n], i = r.parentElement, o = i.getAttribute("data-table-id"), a = ~~i.getAttribute("data-column"),
                l = r.offsetWidth + 1;
            o in t || (t[o] = []);
            for (var s = t[o]; s.length <= a;) s.push(0);
            s[a] < l && (s[a] = l);
        }
        return t;
    }, Tt.prototype.makeCSS = function (e) {
        var t = [];
        for (var n in e) for (var r = e[n], i = "pre.HyperMD-table-row.HyperMD-table_" + n + " .hmd-table-column-", o = 0; o < r.length; o++) {
            var a = r[o];
            t.push("" + i + o + " { min-width: " + (a + .5) + "px }");
        }
        return t.join("\n");
    };
    var xt = x("TableAlign", Tt, yt),
        Lt = Object.freeze({defaultOption: yt, suggestedOption: wt, TableAlign: Tt, getAddon: xt}), Mt = {source: null},
        Ct = {source: "function" == typeof requirejs ? "~codemirror/" : "https://cdn.jsdelivr.net/npm/codemirror/"};
    l.hmdModeLoader = Ct, U.defineOption("hmdModeLoader", Mt, function (e, t) {
        t && "boolean" != typeof t ? "string" != typeof t && "function" != typeof t || (t = {source: t}) : t = {source: t && Ct.source || null};
        var n = Ht(e);
        for (var r in Mt) n[r] = r in t ? t[r] : Mt[r];
    });
    var Ot = function (e) {
        var l = this;
        this.cm = e, this._loadingModes = {}, this.rlHandler = function (e, t) {
            var n = t.lineNo(), r = (t.text || "").match(/^```\s*(\S+)/);
            if (r) {
                var i = r[1], o = U.findModeByName(i), a = o && o.mode;
                !a || a in U.modes || l.startLoadMode(a, n);
            }
        }, (new s).bind(this, "source").ON(function () {
            e.on("renderLine", l.rlHandler);
        }).OFF(function () {
            e.off("renderLine", l.rlHandler);
        });
    };
    Ot.prototype.touchLine = function (e) {
        var t = this.cm.getLineHandle(e), n = t.text.length;
        this.cm.replaceRange(t.text.charAt(n - 1), {line: e, ch: n - 1}, {line: e, ch: n});
    }, Ot.prototype.startLoadMode = function (e, t) {
        var n = this._loadingModes, r = this;
        if (0 <= t && e in n) n[e].push(t); else {
            0 <= t && (n[e] = [t]);
            var i = function () {
                console.log("[HyperMD] mode-loader loaded " + e);
                var t = n[e];
                r.cm.operation(function () {
                    for (var e = 0; e < t.length; e++) r.touchLine(t[e]);
                }), delete n[e];
            }, o = function () {
                console.warn("[HyperMD] mode-loader failed to load mode " + e + " from ", a), -1 !== t && (console.log("[HyperMD] mode-loader will retry loading " + e), setTimeout(function () {
                    r.startLoadMode(e, 0 <= t ? -3 : t + 1);
                }, 1e3));
            };
            if ("function" != typeof this.source) {
                var a = this.source + "mode/" + e + "/" + e + ".js";
                if ("function" == typeof requirejs && "~" === a.charAt(0)) requirejs([a.slice(1, -3)], i); else {
                    var l = document.createElement("script");
                    l.onload = i, l.onerror = o, l.src = a, document.head.appendChild(l);
                }
            } else this.source(e, i, o);
        }
    };
    var Ht = x("ModeLoader", Ot, Mt),
        St = Object.freeze({defaultOption: Mt, suggestedOption: Ct, ModeLoader: Ot, getAddon: Ht});

    function Et(e, t) {
        var n = " " + e.className + " ", r = " " + t + " ";
        return -1 !== n.indexOf(r) && (e.className = n.replace(r, "").trim(), !0);
    }

    function Dt(e, t) {
        var n = " " + t + " ";
        return -1 === (" " + e.className + " ").indexOf(n) && (e.className = e.className + " " + t, !0);
    }

    var Nt = {enabled: !1, line: !0, tokenTypes: "em|strong|strikethrough|code|linkText|task".split("|")},
        At = {enabled: !0};
    l.hmdHideToken = At, c.hmdHideToken = !1, U.defineOption("hmdHideToken", Nt, function (e, t) {
        t && "boolean" != typeof t ? "string" == typeof t ? t = {
            enabled: !0,
            tokenTypes: t.split("|")
        } : t instanceof Array && (t = {enabled: !0, tokenTypes: t}) : t = {enabled: !!t};
        var n = Ft(e);
        for (var r in Nt) n[r] = r in t ? t[r] : Nt[r];
    });
    var Rt = "hmd-hidden-token", It = "hmd-inactive-line", jt = function (e) {
        var r = this;
        this.cm = e, this.renderLineHandler = function (e, t, n) {
            r.procLine(t, n);
        }, this.cursorActivityHandler = function (e) {
            r.update();
        }, this.update = h(function () {
            return r.updateImmediately();
        }, 100), this._rangesInLine = {}, new s(function () {
            e.on("cursorActivity", r.cursorActivityHandler), e.on("renderLine", r.renderLineHandler), e.on("update", r.update), r.update(), e.refresh();
        }, function () {
            e.off("cursorActivity", r.cursorActivityHandler), e.off("renderLine", r.renderLineHandler), e.off("update", r.update), r.update.stop(), e.refresh();
        }).bind(this, "enabled", !0);
    };
    jt.prototype.procLine = function (e, t) {
        var n = this.cm, r = "number" == typeof e ? e : e.lineNo();
        "number" == typeof e && (e = n.getLineHandle(e));
        var i = this._rangesInLine[r] || [], o = b(n, r);
        if (!o || o.hidden || !o.measure) return !1;
        if (t || (t = o.text), !t) return !1;
        var u = _(o, e, r).map, f = u.length / 3, a = !1;

        function l(e, t, n) {
            for (var r = !1, i = n = n || 0; i < f; i++) {
                var o = u[3 * i], a = (u[3 * i + 1], u[3 * i + 2]);
                if (o === e.head.start) {
                    if (/formatting-/.test(e.head.type) && a.nodeType === Node.TEXT_NODE) {
                        var l = a.parentElement;
                        (t ? Dt(l, Rt) : Et(l, Rt)) && (r = !0);
                    }
                    if (e.tail && /formatting-/.test(e.tail.type)) for (var s = i + 1; s < f; s++) {
                        var c = u[3 * s], d = (u[3 * s + 1], u[3 * s + 2]);
                        if (c == e.tail.start && d.nodeType === Node.TEXT_NODE) {
                            var h = d.parentElement;
                            (t ? Dt(h, Rt) : Et(h, Rt)) && (r = !0);
                        }
                        if (c >= e.tail.end) break;
                    }
                }
                if (o >= e.begin) break;
            }
            return r;
        }

        0 === i.length ? Dt(t, It) && (a = !0) : Et(t, It) && (a = !0);
        for (var s = T(n).extract(r), c = 0, d = 0; d < s.length; d++) {
            var h = s[d];
            if (-1 !== this.tokenTypes.indexOf(h.type)) {
                for (var m = [{line: r, ch: h.begin}, {
                    line: r,
                    ch: h.end
                }], p = h.begin; c < f && u[3 * c + 1] < p;) c++;
                for (var g = !0, v = 0; v < i.length; v++) {
                    if (k(m, i[v])) {
                        g = !1;
                        break;
                    }
                }
                l(h, g, c) && (a = !0);
            }
        }
        return a && (delete o.measure.heights, o.measure.cache = {}), a;
    }, jt.prototype.updateImmediately = function () {
        var r = this;
        this.update.stop();
        for (var i = this.cm, e = i.listSelections(), o = {}, a = {}, l = this._rangesInLine, t = 0, n = e; t < n.length; t += 1) {
            var s = g(n[t]), c = s[0].line, d = s[1].line;
            o[c] = o[d] = !0;
            for (var h = c; h <= d; h++) a[h] ? a[h].push(s) : a[h] = [s];
        }
        this._rangesInLine = a, i.operation(function () {
            for (var e in l) a[e] || r.procLine(~~e);
            var t = !1;
            for (var n in a) {
                r.procLine(~~n) && o[n] && (t = !0);
            }
            t && i.refresh();
        });
    };
    var Ft = x("HideToken", jt, Nt),
        Pt = Object.freeze({defaultOption: Nt, suggestedOption: At, HideToken: jt, getAddon: Ft}), qt = {enabled: !1},
        $t = {enabled: !0};
    l.hmdCursorDebounce = $t, U.defineOption("hmdCursorDebounce", qt, function (e, t) {
        t && "boolean" != typeof t || (t = {enabled: !!t});
        var n = Kt(e);
        for (var r in qt) n[r] = r in t ? t[r] : qt[r];
    });
    var zt = function (e) {
            var r = this;
            this.cm = e, this.mouseDownHandler = function (e, t) {
                r.lastX = t.clientX, r.lastY = t.clientY;
                var n = r.mouseMoveSuppress;
                document.addEventListener("mousemove", n, !0), r.lastTimeout && clearTimeout(r.lastTimeout), r.lastTimeout = setTimeout(function () {
                    document.removeEventListener("mousemove", n, !0), r.lastTimeout = null;
                }, 100);
            }, this.mouseMoveSuppress = function (e) {
                Math.abs(e.clientX - r.lastX) <= 5 && Math.abs(e.clientY - r.lastY) <= 5 && e.stopPropagation();
            }, new s(function () {
                e.on("mousedown", r.mouseDownHandler);
            }, function () {
                e.off("mousedown", r.mouseDownHandler);
            }).bind(this, "enabled", !0);
        }, Kt = x("CursorDebounce", zt, qt),
        Ut = Object.freeze({defaultOption: qt, suggestedOption: $t, CursorDebounce: zt, getAddon: Kt}),
        Wt = /^(\s*)(>[> ]*|[*+-] \[[x ]\]\s|[*+-]\s|(\d+)([.)]))(\s*)/,
        Bt = /^(\s*)(>[> ]*|[*+-] \[[x ]\]|[*+-]|(\d+)[.)])(\s*)$/, Xt = /[*+-]\s/,
        Yt = /^(\s*)([*+-]\s|(\d+)([.)]))(\s*)/, Gt = function (e) {
            return /hmd-table-sep/.test(e.type) && !/hmd-table-sep-dummy/.test(e.type);
        };

    function Vt(e) {
        if (e.getOption("disableInput")) return U.Pass;
        for (var t = [], n = 0, r = e.listSelections(); n < r.length; n += 1) {
            var i = r[n], o = i.head, a = i.empty(), l = e.getStateAfter(o.line), s = e.getLine(o.line), c = !1;
            if (!c) {
                var d = !1 !== l.list, h = l.quote, u = Wt.exec(s), f = /^\s*$/.test(s.slice(0, o.ch));
                if (a && (d || h) && u && !f) if (c = !0, Bt.test(s)) />\s*$/.test(s) || e.replaceRange("", {
                    line: o.line,
                    ch: 0
                }, {line: o.line, ch: o.ch + 1}), t.push("\n"); else {
                    var m = u[1], p = u[5], g = !(Xt.test(u[2]) || 0 <= u[2].indexOf(">")),
                        v = g ? parseInt(u[3], 10) + 1 + u[4] : u[2].replace("x", " ");
                    t.push("\n" + m + v + p), g && nn(e, o);
                }
            }
            if (!c) {
                var b = a ? l.hmdTable : 0;
                if (0 != b) {
                    if (!/^[\s\|]+$/.test(s) || o.line !== e.lastLine() && e.getStateAfter(o.line + 1).hmdTable === b) {
                        var _ = $("  |  ", l.hmdTableColumns.length - 1), k = "\n";
                        2 === b && (k += "| ", _ += " |"), 0 == l.hmdTableRow ? e.setCursor({
                            line: o.line + 1,
                            ch: e.getLine(o.line + 1).length
                        }) : e.setCursor({
                            line: o.line,
                            ch: s.length
                        }), e.replaceSelection(k), e.replaceSelection(_, "start");
                    } else e.setCursor({line: o.line, ch: 0}), e.replaceRange("\n", {
                        line: o.line,
                        ch: 0
                    }, {line: o.line, ch: s.length});
                    return void (c = !0);
                }
                if (a && o.ch >= s.length && !l.code && !l.hmdInnerMode && /^\|.+\|.+\|$/.test(s)) {
                    for (var y = e.getLineTokens(o.line), w = "|", T = "|", x = 1; x < y.length; x++) {
                        var L = y[x];
                        "|" !== L.string || L.type && L.type.trim().length || (w += " ------- |", T += "   |");
                    }
                    return e.setCursor({
                        line: o.line,
                        ch: s.length
                    }), e.replaceSelection("\n" + w + "\n| "), e.replaceSelection(T.slice(1) + "\n", "start"), void (c = !0);
                }
            }
            if (c || a && "$$" == s.slice(o.ch - 2, o.ch) && /math-end/.test(e.getTokenTypeAt(o)) && (t.push("\n"), c = !0), !c) return void e.execCommand("newlineAndIndent");
        }
        e.replaceSelections(t);
    }

    function Zt(e) {
        if (e.getOption("disableInput")) return U.Pass;
        for (var t = e.listSelections(), n = a("\n", t.length), r = 0; r < t.length; r++) {
            var i = t[r].head, o = e.getStateAfter(i.line);
            !1 !== o.list && (n[r] += $(" ", o.listStack.slice(-1)[0]));
        }
        e.replaceSelections(n);
    }

    function Qt(e, t, n) {
        if (n && !(n < 0)) {
            var r = /^ */.exec(e.getLine(t))[0].length;
            r < n && (n = r), 0 < n && e.replaceRange("", {line: t, ch: 0}, {line: t, ch: n});
        }
    }

    function Jt(e) {
        for (var t, n = e.listSelections(), r = new z(e), i = 0; i < n.length; i++) {
            var o = n[i], a = o.head, l = o.anchor;
            !o.empty() && 0 < U.cmpPos(a, l) ? (l = (t = [a, l])[0], a = t[1]) : l === a && (l = o.anchor = {
                ch: a.ch,
                line: a.line
            });
            var s = e.getStateAfter(a.line);
            if (s.hmdTable) {
                r.setPos(a.line, a.ch);
                var c = 2 === s.hmdTable, d = a.line, h = e.getLine(d), u = 0, f = 0, m = r.findPrev(Gt);
                if (m) {
                    u = (p = r.findPrev(Gt, m.i_token - 1)) ? p.token.end : 0, f = m.token.start, 0 == u && c && (u += h.match(/^\s*\|/)[0].length);
                } else {
                    if (0 == s.hmdTableRow) return;
                    var p;
                    2 == s.hmdTableRow && d--, d--, h = e.getLine(d), r.setPos(d, h.length), u = (p = r.findPrev(Gt)).token.end, f = h.length, c && (f -= h.match(/\|\s*$/)[0].length);
                }
                return " " === h.charAt(u) && (u += 1), 0 < u && " |" === h.substr(u - 1, 2) && u--, " " === h.charAt(f - 1) && (f -= 1), void e.setSelection({
                    line: d,
                    ch: u
                }, {line: d, ch: f});
            }
            if (0 < s.listStack.length) {
                for (var g = a.line; !Yt.test(e.getLine(g));) {
                    if (g--, !(0 < e.getStateAfter(g).listStack.length)) {
                        g++;
                        break;
                    }
                }
                for (var v = e.lastLine(), b = void 0; g <= l.line && (b = Yt.exec(e.getLine(g))); g++) {
                    var _ = e.getStateAfter(g).listStack, k = _.length, y = 0;
                    for (Qt(e, g, y = 1 == k ? b[1].length : _[k - 1] - (_[k - 2] || 0)); ++g <= v;) {
                        if (e.getStateAfter(g).listStack.length !== k) {
                            g = 1 / 0;
                            break;
                        }
                        if (Yt.test(e.getLine(g))) {
                            g--;
                            break;
                        }
                        Qt(e, g, y);
                    }
                }
                return;
            }
        }
        e.execCommand("indentLess");
    }

    function en(e) {
        var t, n = e.listSelections(), r = [], i = [], o = [], a = {}, l = new z(e), s = !1, c = !1, d = !1, h = !0;

        function u(e) {
            (r[p] = e) && (c = !0);
        }

        function f(e) {
            (i[p] = e) && (d = !0);
        }

        function m(e) {
            (o[p] = e) && (h = !0);
        }

        for (var p = 0; p < n.length; p++) {
            r[p] = i[p] = o[p] = "";
            var g = n[p], v = g.head, b = g.anchor, _ = g.empty();
            !_ && 0 < U.cmpPos(v, b) ? (b = (t = [v, b])[0], v = t[1]) : b === v && (b = g.anchor = {
                ch: v.ch,
                line: v.line
            });
            var k = e.getStateAfter(v.line), y = e.getLine(v.line);
            if (k.hmdTable) {
                s = !0;
                var w = 2 === k.hmdTable, T = k.hmdTableColumns;
                l.setPos(v.line, v.ch);
                var x = l.findNext(Gt, l.i_token);
                if (x) {
                    var L = l.findNext(/hmd-table-sep/, x.i_token + 1);
                    v.ch = x.token.end, b.ch = L ? L.token.start : y.length, b.ch > v.ch && " " === y.charAt(v.ch) && v.ch++, b.ch > v.ch && " " === y.charAt(b.ch - 1) && b.ch--, m(b.ch > v.ch ? e.getRange(v, b) : "");
                } else {
                    var M = 0 === k.hmdTableRow ? 2 : 1;
                    if (v.line + M > e.lastLine() || e.getStateAfter(v.line + M).hmdTable != k.hmdTable) {
                        v.ch = b.ch = y.length;
                        var C = $("  |  ", T.length - 1);
                        0 === k.hmdTableRow && (b.line = v.line += 1, b.ch = v.ch = e.getLine(v.line).length), w ? (u("\n| "), f(C + " |")) : (u("\n"), f(C.trimRight())), m("");
                    } else {
                        b.line = v.line += M, l.setPos(v.line, 0);
                        var O = l.line.text, H = w && l.findNext(/hmd-table-sep-dummy/, 0),
                            S = l.findNext(/hmd-table-sep/, H ? H.i_token + 1 : 1);
                        v.ch = H ? H.token.end : 0, b.ch = S ? S.token.start : O.length, b.ch > v.ch && " " === O.charAt(v.ch) && v.ch++, b.ch > v.ch && " " === O.charAt(b.ch - 1) && b.ch--, m(b.ch > v.ch ? e.getRange(v, b) : "");
                    }
                }
            } else if (0 < k.listStack.length) {
                for (var E = v.line, D = void 0; !(D = Yt.exec(e.getLine(E)));) {
                    if (E--, !(0 < e.getStateAfter(E).listStack.length)) {
                        E++;
                        break;
                    }
                }
                for (var N = e.firstLine(), A = e.lastLine(); E <= b.line && (D = Yt.exec(e.getLine(E))); E++) {
                    var R = e.getStateAfter(E).listStack, I = e.getStateAfter(E - 1).listStack, j = R.length, F = "";
                    for (N < E && j <= I.length && (F = j == I.length ? $(" ", I[j - 1] - D[1].length) : $(" ", I[j] - D[0].length)), a[E] = F; ++E <= A;) {
                        if (e.getStateAfter(E).listStack.length !== j) {
                            E = 1 / 0;
                            break;
                        }
                        if (Yt.test(e.getLine(E))) {
                            E--;
                            break;
                        }
                        a[E] = F;
                    }
                }
                if (!_) {
                    h = !1;
                    break;
                }
            } else if (_) u("    "); else {
                m(e.getRange(v, b));
                for (var P = v.line; P <= b.line; P++) P in a || (a[P] = "    ");
            }
        }
        for (var q in a) a[q] && e.replaceRange(a[q], {line: ~~q, ch: 0});
        s && e.setSelections(n), c && e.replaceSelections(r), d && e.replaceSelections(i, "start"), h && e.replaceSelections(o, "around");
    }

    function tn(T, x, L) {
        return function (e) {
            var t;
            if (e.getOption("disableInput")) return U.Pass;
            for (var n = new z(e), r = e.listSelections(), i = new Array(r.length), o = 0; o < r.length; o++) {
                var a = r[o], l = a.head, s = a.anchor, c = e.getStateAfter(l.line), d = a.empty();
                0 < U.cmpPos(l, s) && (s = (t = [l, s])[0], l = t[1]);
                var h = i[o] = d ? "" : e.getRange(l, s);
                if (d || T(e.getTokenAt(l).state)) {
                    var u = l.line;
                    n.setPos(u, l.ch, !0);
                    var f = n.lineTokens[n.i_token];
                    f && f.state;
                    f && !/^\s*$/.test(f.string) || (f = n.lineTokens[--n.i_token]);
                    var m = n.expandRange(function (e) {
                        return e && (T(e.state) || x(e));
                    }), p = m.from, g = m.to;
                    if (g.i_token === p.i_token) {
                        var v = L();
                        if (f && !/^\s*$/.test(f.string)) {
                            var b = {line: u, ch: f.start}, _ = {line: u, ch: f.end};
                            return f = p.token, e.replaceRange(v + f.string + v, b, _), _.ch += v.length, void e.setCursor(_);
                        }
                        i[o] = v;
                    } else x(g.token) && e.replaceRange("", {line: u, ch: g.token.start}, {
                        line: u,
                        ch: g.token.end
                    }), p.i_token !== g.i_token && x(p.token) && e.replaceRange("", {
                        line: u,
                        ch: p.token.start
                    }, {line: u, ch: p.token.end});
                } else {
                    var k = e.getTokenAt(l), y = k ? k.state : c, w = L(y);
                    i[o] = w + h + w;
                }
            }
            e.replaceSelections(i);
        };
    }

    function nn(e, t) {
        var n = Wt, r = t.line, i = 0, o = 0, a = n.exec(e.getLine(r)), l = a[1];
        do {
            var s = r + (i += 1), c = e.getLine(s), d = n.exec(c);
            if (d) {
                var h = d[1], u = parseInt(a[3], 10) + i - o, f = parseInt(d[3], 10), m = f;
                if (l !== h || isNaN(f)) {
                    if (l.length > h.length) return;
                    if (l.length < h.length && 1 === i) return;
                    o += 1;
                } else u === f && (m = f + 1), f < u && (m = u + 1), e.replaceRange(c.replace(n, h + m + d[4] + d[5]), {
                    line: s,
                    ch: 0
                }, {line: s, ch: c.length});
            }
        } while (d);
    }

    Object.assign(U.commands, {hmdNewlineAndContinue: Vt, hmdNewline: Zt, hmdShiftTab: Jt, hmdTab: en});
    var rn = U.keyMap.default === U.keyMap.macDefault ? "Cmd" : "Ctrl",
        on = {"Shift-Tab": "hmdShiftTab", Tab: "hmdTab", Enter: "hmdNewlineAndContinue", "Shift-Enter": "hmdNewline"};
    on[rn + "-B"] = tn(function (e) {
        return e.strong;
    }, function (e) {
        return / formatting-strong /.test(e.type);
    }, function (e) {
        return $(e && e.strong || "*", 2);
    }), on[rn + "-I"] = tn(function (e) {
        return e.em;
    }, function (e) {
        return / formatting-em /.test(e.type);
    }, function (e) {
        return e && e.em || "*";
    }), on[rn + "-D"] = tn(function (e) {
        return e.strikethrough;
    }, function (e) {
        return / formatting-strikethrough /.test(e.type);
    }, function (e) {
        return "~~";
    }), on.fallthrough = "default", on = U.normalizeKeyMap(on), U.keyMap.hypermd = on, l.keyMap = "hypermd";
    var an = Object.freeze({
        newlineAndContinue: Vt, newline: Zt, shiftTab: Jt, tab: en, wrapTexts: function (e, t, n) {
            var r;
            if (e.getOption("disableInput")) return U.Pass;
            var i = e.listSelections(), o = new Array(i.length), a = new Array(i.length), l = !1, s = !1, c = !1;
            n || (n = t);
            for (var d = t.length, h = n.length, u = 0; u < i.length; u++) {
                o[u] = a[u] = "";
                var f = i[u], m = f.head, p = f.anchor, g = e.getLine(m.line);
                if (f.empty()) m.ch >= d && g.substr(m.ch - d, d) === t ? (c = !0, m.ch -= d) : (s = !0, a[u] = t); else {
                    l = !0, 0 < U.cmpPos(m, p) && (p = (r = [m, p])[0], m = r[1]);
                    var v = e.getRange(m, p);
                    m.ch >= d && m.line === p.line && g.substr(m.ch - d, d) === t && g.substr(p.ch, h) === n && (c = !0, p.ch += h, m.ch -= d, v = t + v + n), v.slice(0, d) === t && v.slice(-h) === n ? o[u] = v.slice(d, -h) : o[u] = t + v + n;
                }
            }
            c && e.setSelections(i), s && e.replaceSelections(a), l && e.replaceSelections(o, "around");
        }, createStyleToggler: tn, get keyMap() {
            return on;
        }
    });
    e.cmpPos = U.cmpPos, e.Mode = C, e.InsertFile = D, e.ReadLink = q, e.Hover = de, e.Click = ve, e.Paste = xe, e.Fold = Ne, e.FoldImage = Re, e.FoldLink = je, e.FoldCode = Ye, e.FoldMath = nt, e.FoldEmoji = ht, e.FoldHTML = kt, e.TableAlign = Lt, e.ModeLoader = St, e.HideToken = Pt, e.CursorDebounce = Ut, e.KeyMap = an, e.Addon = L, e.FlipFlop = s, e.tryToRun = d, e.debounce = h, e.addClass = t, e.rmClass = n, e.contains = r, e.repeat = a, e.repeatStr = $, e.visitElements = m, e.watchSize = p, e.makeSymbol = i, e.suggestedEditorConfig = l, e.normalVisualConfig = c, e.fromTextArea = function (e, t) {
        var n = Object.assign({}, l, t), r = U.fromTextArea(e, n);
        return r[o] = !0, r;
    }, e.switchToNormal = function (e, t) {
        if (e[o]) {
            "string" == typeof t && (t = {theme: t});
            var n = Object.assign({}, c, t);
            for (var r in n) e.setOption(r, n[r]);
        }
    }, e.switchToHyperMD = function (e, t) {
        "string" == typeof t && (t = {theme: t});
        var n = {};
        if (o in e) {
            for (var r in c) n[r] = l[r];
            Object.assign(n, t);
        } else Object.assign(n, l, t), e[o] = !0;
        for (var i in n) e.setOption(i, n[i]);
    }, e.cm_internal = f, e.TokenSeeker = z, e.getEveryCharToken = function (e) {
        var t = new Array(e.text.length), n = e.styles, r = 0;
        if (n) for (var i = 1; i < n.length; i += 2) for (var o = n[i], a = n[i + 1]; r < o;) t[r++] = a; else for (var l = (e.parent.cm || e.parent.parent.cm || e.parent.parent.parent.cm).getLineTokens(e.lineNo()), s = 0; s < l.length; s++) for (var c = l[s].end, d = l[s].type; r < c;) t[r++] = d;
        return t;
    }, e.expandRange = w, e.orderedRange = g, e.rangesIntersect = k, e.getLineSpanExtractor = T, Object.defineProperty(e, "__esModule", {value: !0});
});
