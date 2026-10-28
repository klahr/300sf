/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Toast.ts"/>
/// <reference path="ui/ConfirmBox.ts"/>
/// <reference path="HttpRequest.ts"/>
/// <reference path="App.ts"/>
/// <reference path="models/Competition.ts"/>

namespace Competitions {
	function locale(): string {
		return App.language === "sv" ? "sv-SE" : App.language === "de" ? "de-DE" : "en-GB";
	}

	// The competition ends at midnight after its last day, so that day is the one before.
	export function period(c: Models.Competition): string {
		const from = new Date(c.from);
		const last = new Date(new Date(c.until).getTime() - 1);
		const sameYear = from.getFullYear() === last.getFullYear();
		const f = from.toLocaleDateString(locale(), sameYear ? { day: "numeric", month: "short" } : { day: "numeric", month: "short", year: "numeric" });
		const l = last.toLocaleDateString(locale(), { day: "numeric", month: "short", year: "numeric" });
		return f + " – " + l;
	}

	export function statusLabel(c: Models.Competition): string {
		return c.status === "upcoming" ? App.translations.CompetitionUpcoming : c.status === "finished" ? App.translations.CompetitionFinished : App.translations.CompetitionOngoing;
	}

	export function link(c: Models.Competition): string {
		return window.location.origin + "/competitions?c=" + encodeURIComponent(c.uuid);
	}

	// What the bars in the standings measure against: the winning count, the list, or the leader.
	function target(c: Models.Competition): number {
		if (c.end_on_reports) {
			return c.end_on_reports;
		}
		const leader = c.standings && c.standings.length > 0 ? c.standings[0].count : 0;
		return Math.max(1, leader);
	}

	function rule(parent: HTMLElement, icon: string, text: string): void {
		const row = document.createElement("div");
		row.className = "competition-rule";
		const i = document.createElement("span");
		i.className = "material-symbols-outlined";
		i.textContent = icon;
		const t = document.createElement("span");
		t.textContent = text;
		row.appendChild(i);
		row.appendChild(t);
		parent.appendChild(row);
	}

	export function standingsList(c: Models.Competition, limit: number = 0): HTMLElement {
		const list = document.createElement("ol");
		list.className = "standings";
		const goal = target(c);
		const rows = limit > 0 ? (c.standings || []).filter((s, i) => i < limit || s.me) : (c.standings || []);
		rows.forEach(s => {
			const li = document.createElement("li");
			if (s.me) {
				li.classList.add("me");
			}
			if (s.rank === 1 && c.status === "finished" && s.count > 0) {
				li.classList.add("winner");
			}
			const rank = document.createElement("span");
			rank.className = "standing-rank";
			rank.textContent = String(s.rank);
			const img = document.createElement("img");
			img.alt = "";
			img.src = "/img/portraits/" + s.user.portrait;
			const who = document.createElement("span");
			who.className = "standing-who";
			const name = document.createElement("b");
			name.textContent = s.me ? App.translations.You : s.user.display_name;
			const handle = document.createElement("small");
			handle.textContent = "@" + s.user.username;
			who.appendChild(name);
			who.appendChild(handle);
			const bar = document.createElement("span");
			bar.className = "standing-bar";
			const fill = document.createElement("span");
			fill.style.width = Math.min(100, s.count / goal * 100) + "%";
			bar.appendChild(fill);
			const count = document.createElement("span");
			count.className = "standing-count";
			count.textContent = String(s.count);
			li.appendChild(rank);
			li.appendChild(img);
			li.appendChild(who);
			li.appendChild(bar);
			li.appendChild(count);
			list.appendChild(li);
		});
		return list;
	}

	function share(c: Models.Competition): void {
		const url = link(c);
		const nav = navigator as any;
		if (nav.share) {
			nav.share({ title: App.getTranslation(c.name), url: url }).catch(function () { });
			return;
		}
		navigator.clipboard.writeText(url).then(function () {
			new Ui.Toast(App.translations.LinkCopied, Utils.escapeHtml(url), null, "var(--success-color)", function () { });
		});
	}

	function invite(c: Models.Competition): void {
		const sheet = new Ui.Sheet({ title: App.translations.InviteFriends, size: "compact" });
		const section = sheet.section();
		const hint = document.createElement("p");
		hint.className = "details";
		hint.textContent = App.translations.InviteHint;
		section.appendChild(hint);
		const input = document.createElement("input");
		input.type = "text";
		input.id = "competition-invite";
		input.placeholder = "@" + App.translations.Username.toLowerCase();
		input.autocapitalize = "off";
		input.spellcheck = false;
		sheet.field(section, App.translations.Usernames, input);
		const result = document.createElement("p");
		result.className = "details";
		section.appendChild(result);

		const send = sheet.addAction(Ui.actionButton(App.translations.SendInvitations, "send", "success"));
		send.disabled = true;
		input.addEventListener("input", function () {
			send.disabled = input.value.trim() === "";
		});
		send.addEventListener("click", function () {
			const usernames = input.value.split(/[\s,;]+/).filter(u => u !== "");
			Ui.setBusy(send, true);
			new HttpRequest("PUT", "/v1/competitions/" + c.uuid + "/invite", { usernames: usernames }).process(function (status: any, data: any) {
				const notFound = (data.not_found || []) as Array<string>;
				if (notFound.length === 0) {
					sheet.close();
					new Ui.Toast(App.translations.InvitationsSent, Utils.escapeHtml((data.invited || []).map((u: string) => "@" + u).join(", ")), null, "var(--success-color)", function () { });
					return;
				}
				input.value = notFound.join(", ");
				result.textContent = App.translations.UsersNotFound.replace("{0}", notFound.map(u => "@" + u).join(", "));
			}, function () {
				result.textContent = App.translations.SomethingWentWrong;
			}, function () {
				Ui.setBusy(send, false);
			});
		});
		sheet.open();
		input.focus();
	}

	// The sheet for one competition. changed is called after joining, leaving or removing it.
	export function open(uuid: string, changed: () => void = function () { }): void {
		const sheet = new Ui.Sheet({ title: App.translations.Competitions, eyebrow: " ", size: "tall" });
		sheet.root.addClass("competition-sheet");
		const loading = sheet.section();
		loading.classList.add("is-loading");
		loading.innerHTML = "<div class='skeleton competition-skeleton'></div><div class='skeleton competition-skeleton'></div><div class='skeleton competition-skeleton tall'></div>";
		sheet.open();

		const load = function () {
			new HttpRequest("GET", "/v1/competitions/" + encodeURIComponent(uuid)).process(function (status: any, data: any) {
				render(data as Models.Competition);
			}, function () {
				sheet.body.innerHTML = "";
				const empty = document.createElement("div");
				empty.className = "empty-state";
				empty.innerHTML = "<span class='material-symbols-outlined'>search_off</span>";
				const p = document.createElement("p");
				p.textContent = App.translations.CompetitionNotFound;
				empty.appendChild(p);
				sheet.section().appendChild(empty);
			});
		};

		const render = function (c: Models.Competition) {
			sheet.body.innerHTML = "";
			sheet.footer.innerHTML = "";
			sheet.footer.classList.add("hidden");
			sheet.title.textContent = App.getTranslation(c.name);
			const eyebrow = sheet.names.querySelector(".sheet-eyebrow");
			if (eyebrow) {
				eyebrow.textContent = statusLabel(c) + " · " + period(c);
			}
			sheet.root.container.style.setProperty("--competition-color", c.color || "var(--primary)");

			// Where the viewer stands, or who won.
			const summary = sheet.section();
			summary.classList.add("competition-summary");
			const winner = c.standings && c.status === "finished" && c.standings.length > 0 && c.standings[0].count > 0 ? c.standings[0] : null;
			if (winner) {
				summary.innerHTML = "<span class='material-symbols-outlined'>emoji_events</span>";
				const t = document.createElement("p");
				t.textContent = App.translations.CompetitionWonBy.replace("{0}", winner.me ? App.translations.You : winner.user.display_name).replace("{1}", String(winner.count));
				summary.appendChild(t);
			} else if (c.joined) {
				summary.innerHTML = "<span class='competition-rank'>" + (c.my_count > 0 ? "#" + c.my_rank : "–") + "</span>";
				const t = document.createElement("p");
				t.textContent = App.translations.CompetitionYourPlace.replace("{0}", String(c.my_count)).replace("{1}", String(c.participants));
				summary.appendChild(t);
			} else {
				summary.remove();
			}

			const description = App.getTranslation(c.description);
			if (description) {
				const d = sheet.section();
				const p = document.createElement("p");
				p.className = "competition-description";
				p.textContent = description;
				d.appendChild(p);
			}

			const rules = sheet.section(App.translations.Rules);
			rule(rules, "event", period(c));
			rule(rules, "checklist", c.list_name ? App.getTranslation(c.list_name) + " · " + App.translations.NumberOfSpecies.replace("{0}", String(c.list_size)) : App.translations.AllSpecies);
			rule(rules, "location_on", c.zones.length > 0 ? c.zones.map(z => App.getTranslation(z)).join(", ") : App.translations.AllowedAnywhereShort);
			rule(rules, "flag", c.end_on_reports ? App.translations.FirstToWins.replace("{0}", String(c.end_on_reports)) : App.translations.MostSpeciesWins);
			rule(rules, "group", c.max_competitors ? App.translations.ParticipantsOfMax.replace("{0}", String(c.participants)).replace("{1}", String(c.max_competitors)) : App.translations.ParticipantsCount.replace("{0}", String(c.participants)));
			rule(rules, c.visibility === Models.Visibility.Public ? "public" : "lock", (c.visibility === Models.Visibility.Public ? App.translations.CompetitionPublic : App.translations.CompetitionPrivate) + (c.has_password ? " · " + App.translations.PasswordRequired : ""));
			if (c.creator && c.creator.username) {
				rule(rules, "person", App.translations.CreatedBy.replace("{0}", c.creator.display_name));
			}

			const standings = sheet.section(App.translations.Standings);
			if (c.standings === null) {
				const p = document.createElement("p");
				p.className = "details";
				p.textContent = App.translations.JoinToSeeStandings;
				standings.appendChild(p);
			} else if (c.standings.length === 0) {
				const p = document.createElement("p");
				p.className = "details";
				p.textContent = App.translations.NoParticipantsYet;
				standings.appendChild(p);
			} else {
				standings.appendChild(standingsList(c));
				const note = document.createElement("p");
				note.className = "details";
				note.textContent = App.translations.StandingsNote;
				standings.appendChild(note);
			}

			// Joining, with the password when one is needed.
			let password: HTMLInputElement | null = null;
			let passwordError: HTMLParagraphElement | null = null;
			if (c.can_join && c.has_password && !c.invited) {
				const join = sheet.section();
				password = document.createElement("input");
				password.type = "password";
				password.id = "competition-password";
				password.autocomplete = "off";
				sheet.field(join, App.translations.Password, password);
				passwordError = document.createElement("p");
				passwordError.className = "details error";
				join.appendChild(passwordError);
			}
			if (c.can_join || c.invited) {
				const joinButton = sheet.addAction(Ui.actionButton(c.invited ? App.translations.AcceptInvitation : App.translations.JoinCompetition, "how_to_reg", "success"));
				joinButton.disabled = !c.can_join;
				joinButton.addEventListener("click", function () {
					Ui.setBusy(joinButton, true);
					new HttpRequest("PUT", "/v1/competitions/" + c.uuid + "/join", { password: password ? password.value : "" }).process(function () {
						changed();
						load();
					}, function (status: any, data: any) {
						if (passwordError) {
							passwordError.textContent = data && data.code === 9001 ? App.translations.WrongPassword : App.translations.SomethingWentWrong;
						} else {
							new Ui.Toast(App.translations.Error, App.translations.SomethingWentWrong, null, "var(--error-color)", function () { });
						}
					}, function () {
						Ui.setBusy(joinButton, false);
					});
				});
				if (password) {
					joinButton.disabled = true;
					password.addEventListener("input", function () {
						joinButton.disabled = password!.value === "";
					});
				}
			}
			if (c.invited) {
				sheet.addAction(Ui.actionButton(App.translations.Decline, "close", "secondary")).addEventListener("click", function () {
					new HttpRequest("PUT", "/v1/competitions/" + c.uuid + "/decline").process(function () {
						changed();
						sheet.close();
					});
				});
			}
			if ((c.joined || c.is_creator) && c.status !== "finished") {
				sheet.addAction(Ui.actionButton(App.translations.Invite, "person_add", "secondary")).addEventListener("click", function () {
					invite(c);
				});
			}
			if (c.visibility === Models.Visibility.Public || c.joined || c.is_creator) {
				sheet.addAction(Ui.actionButton(App.translations.Share, "share", "secondary")).addEventListener("click", function () {
					share(c);
				});
			}

			// Leaving and removing, kept apart from the everyday actions.
			if (c.joined || c.is_creator) {
				const more = sheet.section();
				more.classList.add("competition-more");
				if (c.joined) {
					const leave = document.createElement("button");
					leave.className = "text-button";
					leave.textContent = App.translations.LeaveCompetition;
					leave.addEventListener("click", function () {
						new Ui.ConfirmBox(App.translations.LeaveCompetition, App.translations.LeaveCompetitionConfirm, null, "var(--error-color)", function () {
							new HttpRequest("PUT", "/v1/competitions/" + c.uuid + "/leave").process(function () {
								changed();
								load();
							});
						}, function () { }, App.translations.LeaveCompetition, App.translations.Cancel);
					});
					more.appendChild(leave);
				}
				if (c.is_creator) {
					const remove = document.createElement("button");
					remove.className = "text-button is-danger";
					remove.textContent = App.translations.RemoveCompetition;
					remove.addEventListener("click", function () {
						new Ui.ConfirmBox(App.translations.RemoveCompetition, App.translations.RemoveCompetitionConfirm, null, "var(--error-color)", function () {
							new HttpRequest("PUT", "/v1/competitions/" + c.uuid + "/archive").process(function () {
								changed();
								sheet.close();
							});
						}, function () { }, App.translations.RemoveCompetition, App.translations.Cancel);
					});
					more.appendChild(remove);
				}
			}
		};

		load();
	}
}
