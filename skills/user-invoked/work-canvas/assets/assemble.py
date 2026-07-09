#!/usr/bin/env python3
"""work-canvas assembler — turn a body file full of paste-markers into ONE
self-contained .html, WITHOUT the model ever reading or re-emitting the assets.

The model writes only the body (the content). This script inlines + minifies
base.css and interactions.js and embeds local images as data: URIs. The final
file is byte-for-byte the kind of self-contained page the skill requires, but
the model spends ~0 tokens on the ~24KB of CSS/JS it would otherwise retype.

Markers (put them in the body where the assets go):
  [[PASTE base.css]]         -> minified base.css   (inside your <style>…</style>)
  [[PASTE interactions.js]]  -> minified interactions.js (inside your <script>…</script>)
  [[IMG path]] or [[IMG:path]] -> data: URI of the image at `path`
                                  (resolved relative to the body file, then CWD)

Usage:
  python assemble.py body.html out.html [--no-minify]

Assets resolve relative to this script, so run it from anywhere.
"""
import base64, mimetypes, os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))

def minify_css(s):
    s = re.sub(r"/\*.*?\*/", "", s, flags=re.S)        # block comments
    s = "\n".join(l.strip() for l in s.splitlines())   # de-indent + de-trail (safe: no multiline strings)
    return re.sub(r"\n{2,}", "\n", s).strip()

def minify_js(s):
    s = re.sub(r"/\*.*?\*/", "", s, flags=re.S)        # block comments (no template literals -> safe)
    out = []
    for l in s.splitlines():
        t = l.strip()
        if not t or t.startswith("//"):
            continue
        out.append(t)                                  # de-indent; keep line breaks (ASI-safe)
    return "\n".join(out)

def datauri(path):
    mime = mimetypes.guess_type(path)[0] or "application/octet-stream"
    return f"data:{mime};base64," + base64.b64encode(open(path, "rb").read()).decode()

def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    minify = "--no-minify" not in sys.argv
    if len(args) < 1:
        sys.exit("usage: python assemble.py body.html [out.html] [--no-minify]")
    body_path = args[0]
    out_path = args[1] if len(args) > 1 else os.path.splitext(body_path)[0] + ".final.html"
    body_dir = os.path.dirname(os.path.abspath(body_path))
    html = open(body_path, encoding="utf-8").read()
    # Drop HTML comments first — models often copy the starter's instructional comment
    # (which contains the literal marker examples), which would otherwise double-inline the
    # CSS and trip on the example [[IMG path]]. Real markers live in <style>/<script>, not comments.
    html = re.sub(r"<!--.*?-->", "", html, flags=re.S)

    css = open(os.path.join(HERE, "base.css"), encoding="utf-8").read()
    js = open(os.path.join(HERE, "interactions.js"), encoding="utf-8").read()
    if minify:
        css, js = minify_css(css), minify_js(js)
    html = html.replace("[[PASTE base.css]]", css).replace("[[PASTE interactions.js]]", js)

    def img(m):
        p = m.group(1).strip()
        for cand in (p, os.path.join(body_dir, p), os.path.join(os.getcwd(), p)):
            if os.path.exists(cand):
                return datauri(cand)
        print(f"  WARN image not found: {p}", file=sys.stderr)
        return m.group(0)
    html = re.sub(r"\[\[IMG:?\s*([^\]]+)\]\]", img, html)

    with open(out_path, "w", encoding="utf-8") as f:
        f.write(html)
    left = [m for m in ("[[PASTE base.css]]", "[[PASTE interactions.js]]") if m in html] \
        + (["[[IMG …]]"] if "[[IMG" in html else [])
    print(f"wrote {out_path} ({len(html.encode())} bytes)" + (f"  UNRESOLVED: {left}" if left else "  ✓"))

if __name__ == "__main__":
    main()
