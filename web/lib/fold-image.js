// HyperMD, copyright (c) by laobubu
// Distributed under an MIT license: http://laobubu.net/HyperMD/LICENSE
//
// DESCRIPTION: Fold Image Markers `![](xxx)`
//

(function (mod){ //[HyperMD] UMD patched!
  /*commonjs*/  ("object"==typeof exports&&"undefined"!=typeof module) ? mod(null, exports, require("./fold"), require("./read-link")) :
  /*amd*/       ("function"==typeof define&&define.amd) ? define(["require","exports","./fold","./read-link"], mod) :
  /*plain env*/ mod(null, (this.HyperMD.FoldImage = this.HyperMD.FoldImage || {}), HyperMD.Fold, HyperMD.ReadLink);
})(function (require, exports, fold_1) {
    "use strict";
    Object.defineProperty(exports, "__esModule", { value: true });
    var DEBUG = true;
    exports.ImageFolder = function (stream, token) {
        var cm = stream.cm;
        var imgRE = /\bimage-marker\b/;
        var urlRE = /\bformatting-link-string\b/; // matches the parentheses
        if (imgRE.test(token.type) && token.string === "!") {
            var lineNo = stream.lineNo;
            // find the begin and end of url part
            var url_begin = stream.findNext(urlRE);
            var url_end = stream.findNext(urlRE, url_begin.i_token + 1);
            var from = { line: lineNo, ch: token.start };
            var to = { line: lineNo, ch: url_end.token.end };

            var rngReq = stream.requestRange(from, to, from, from);
            if (rngReq === fold_1.RequestRangeResult.OK) {
                // That fixes blinking on select, for some reason range CI is returned even though cursor is outside of our tokens
                if (cm.somethingSelected()) {
                    return null;
                }

                var url;
                var title;
                { // extract the URL
                    var rawurl = cm.getRange(// get the URL or footnote name in the parentheses
                    { line: lineNo, ch: url_begin.token.start + 1 }, { line: lineNo, ch: url_end.token.start });
                    if (url_end.token.string === "]") {
                        var tmp = cm.hmdReadLink(rawurl, lineNo);
                        if (!tmp)
                            return null; // Yup! bad URL?!
                        rawurl = tmp.content;
                    }
                    url = cm.hmdResolveURL(rawurl);
                }
                { // extract the title
                    title = cm.getRange({ line: lineNo, ch: from.ch + 2 }, { line: lineNo, ch: url_begin.token.start - 1 });
                }
                let img = document.createElement("img");
                img.style.cursor = "pointer";
                // PATCHED, we don't want blank line with the cursor after image
                let wrapper = document.createElement("span");
                wrapper.style.display = "inline-flex";
                wrapper.style.justifyContent = "center";
                wrapper.style.alignItems = "center";
                wrapper.style.width = "100%";
                wrapper.style.textAlign = "center";
                wrapper.appendChild(img);
                wrapper.addEventListener('click', function () {
                    cm.focus();
                    const lineNo = from.line;
                    const lineLength = cm.getLine(lineNo).length;
                    cm.setCursor({ line: lineNo, ch: lineLength });
                });

                var marker = cm.markText(from, to, {
                    clearOnEnter: false,
                    collapsed: true,
                    // PATCHED, was img
                    replacedWith: img,
                });
                img.addEventListener('click', function (e) {
                    e.stopPropagation();
                    let modal = document.createElement("div");
                    modal.style.position = "fixed";
                    modal.style.top = "0";
                    modal.style.left = "0";
                    modal.style.width = "100vw";
                    modal.style.height = "100vh";
                    modal.style.backgroundColor = "rgba(0, 0, 0, 0.8)";
                    modal.style.display = "flex";
                    modal.style.justifyContent = "center";
                    modal.style.alignItems = "center";
                    modal.style.zIndex = "1000";

                    let imgPreview = document.createElement("img");
                    imgPreview.src = img.src;
                    imgPreview.className = "hmd-image-preview";
                    imgPreview.style.maxWidth = "90%";
                    imgPreview.style.maxHeight = "90%";
                    imgPreview.style.borderRadius = "8px";

                    modal.appendChild(imgPreview);

                    const closeModal = () => {
                        document.body.removeChild(modal);
                        document.removeEventListener("keydown", handleKeyDown);
                    };

                    modal.addEventListener("click", closeModal);
                    const handleKeyDown = (event) => {
                        if (event.key === "Escape") {
                            closeModal();
                        }
                    };
                    document.addEventListener("keydown", handleKeyDown);

                    document.body.appendChild(modal);
                }, false);
                img.addEventListener('load', function () {
                    img.classList.remove("hmd-image-loading");
                    marker.changed();
                }, false);
                img.addEventListener('error', function () {
                    img.classList.remove("hmd-image-loading");
                    img.classList.add("hmd-image-error");
                    marker.changed();
                }, false);
                // img.addEventListener('click', function () { return fold_1.breakMark(cm, marker); }, false);
                img.className = "hmd-image hmd-image-loading";
                img.src = url;
                img.title = title;
                return marker;
            }
            else {
                if (DEBUG) {
                    console.log("[image]FAILED TO REQUEST RANGE: ", rngReq);
                }
            }
        }
        return null;
    };
    fold_1.registerFolder("image", exports.ImageFolder, true);
});