/// <reference path="Html.ts"/>

// The text of news posts. Admins write a small part of Markdown:
//
//   # Heading            a line starting with "# "
//   - item               lines starting with "- " make a list
//   **bold**, *italic*, [text](https://link)
//   a blank line         starts a new paragraph
//
// Posts from before were written in HTML; those keep only harmless tags.
// Either way nothing in a post can run a script or style the page.
namespace NewsText {
	export function localized(t: Models.LocalizedText | null | undefined): string {
		if (!t) {
			return "";
		}
		const own = App.language === "sv" ? t.sv : App.language === "de" ? t.de : t.en;
		return own || t.en || "";
	}

	export function render(text: string): DocumentFragment {
		const template = document.createElement("template");
		template.innerHTML = /<\/?[a-z][^>]*>/i.test(text) ? sanitize(text) : markdown(text);
		return template.content;
	}

	// The first paragraph as plain text, for previews.
	export function summary(text: string, max = 160): string {
		const plain = (render(text).textContent || "").replace(/\s+/g, " ").trim();
		return plain.length > max ? plain.slice(0, max - 1).replace(/\s+$/, "") + "…" : plain;
	}

	function inline(text: string): string {
		let out = "";
		const link = /\[([^\]]+)\]\(([^)\s]+)\)/g;
		let last = 0;
		let m: RegExpExecArray | null;
		while ((m = link.exec(text)) !== null) {
			out += emphasis(Utils.escapeHtml(text.slice(last, m.index)));
			const href = Utils.safeUrl(m[2]);
			const label = emphasis(Utils.escapeHtml(m[1]));
			out += href ? "<a href=\"" + Utils.escapeHtml(href) + "\" target=\"_blank\" rel=\"noopener\">" + label + "</a>" : label;
			last = m.index + m[0].length;
		}
		return out + emphasis(Utils.escapeHtml(text.slice(last)));
	}

	// On escaped text, so * is the only markup left to find.
	function emphasis(escaped: string): string {
		return escaped
			.replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
			.replace(/\*(.+?)\*/g, "<em>$1</em>");
	}

	function markdown(text: string): string {
		return text.replace(/\r\n?/g, "\n").split(/\n\s*\n/).map(block => {
			const lines = block.split("\n").filter(l => l.trim() !== "");
			if (lines.length === 0) {
				return "";
			}
			if (lines.every(l => /^\s*[-*] /.test(l))) {
				return "<ul>" + lines.map(l => "<li>" + inline(l.replace(/^\s*[-*] /, "")) + "</li>").join("") + "</ul>";
			}
			if (lines.length === 1 && /^#{1,3} /.test(lines[0])) {
				return "<h3>" + inline(lines[0].replace(/^#{1,3} /, "")) + "</h3>";
			}
			return "<p>" + lines.map(inline).join("<br>") + "</p>";
		}).join("");
	}

	const allowed = ["P", "BR", "B", "STRONG", "I", "EM", "A", "UL", "OL", "LI", "H2", "H3", "H4"];

	// Rebuilds the HTML from allowed tags only: others are replaced by their
	// text, and no attribute survives but a link's http(s) address.
	function sanitize(html: string): string {
		const source = new DOMParser().parseFromString(html, "text/html").body;
		const out = document.createElement("div");
		const copy = function (from: Node, to: Node) {
			from.childNodes.forEach(n => {
				if (n.nodeType === Node.TEXT_NODE) {
					to.appendChild(document.createTextNode(n.textContent || ""));
				} else if (n instanceof Element) {
					if (n.tagName === "SCRIPT" || n.tagName === "STYLE") {
						return;
					}
					if (allowed.indexOf(n.tagName) < 0) {
						copy(n, to);
						return;
					}
					const e = document.createElement(n.tagName === "H2" ? "h3" : n.tagName.toLowerCase());
					if (n.tagName === "A") {
						const href = Utils.safeUrl(n.getAttribute("href"));
						if (href) {
							e.setAttribute("href", href);
							e.setAttribute("target", "_blank");
							e.setAttribute("rel", "noopener");
						}
					}
					copy(n, e);
					to.appendChild(e);
				}
			});
		};
		copy(source, out);
		return out.innerHTML;
	}
}
