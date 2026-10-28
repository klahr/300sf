/// <reference path="ui/Sheet.ts"/>
/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>
/// <reference path="HttpRequest.ts"/>

// Records up to ten seconds of a bird and asks the identification service (BirdNET, behind
// /api/v1/analyze) what it is. Recorded as plain WAV in the browser, which the service can always
// read, so nothing has to be loaded for it. Picking a result calls pick with the species.
class AnalyzePopUp {
	private static readonly SECONDS = 10;

	private mSheet: Ui.Sheet;
	private mButton: HTMLButtonElement;
	private mStatus: HTMLParagraphElement;
	private mResults: HTMLElement;
	private mStream: MediaStream | null = null;
	private mContext: AudioContext | null = null;
	private mSamples: Array<Float32Array> = [];
	private mTimer: number | null = null;
	private mRecording = false;
	private mClosed = false;
	private mPick: (species: Models.Species) => void;

	constructor(pick: (species: Models.Species) => void) {
		const self = this;
		this.mPick = pick;

		this.mSheet = new Ui.Sheet({ title: App.translations.IdentifyBySound, size: "auto" });

		const record = this.mSheet.section();
		record.classList.add("identify");

		const description = document.createElement("p");
		description.className = "identify-description";
		description.textContent = App.translations.AnalyzeDescription;
		record.appendChild(description);

		this.mButton = document.createElement("button");
		this.mButton.className = "identify-button";
		this.mButton.setAttribute("aria-label", App.translations.AnalyzeStartRecording);
		this.mButton.innerHTML = "<svg viewBox='0 0 120 120' aria-hidden='true'><circle class='track' cx='60' cy='60' r='56'/><circle class='progress' cx='60' cy='60' r='56'/></svg><span class='material-symbols-outlined'>mic</span>";
		this.mButton.addEventListener("click", function () {
			if (self.mRecording) {
				self.stop();
			} else {
				self.start();
			}
		});
		record.appendChild(this.mButton);

		this.mStatus = document.createElement("p");
		this.mStatus.className = "identify-status";
		this.mStatus.textContent = App.translations.AnalyzeStartRecording;
		this.mStatus.setAttribute("role", "status");
		record.appendChild(this.mStatus);

		this.mResults = this.mSheet.section(App.translations.PossibleSpecies);
		this.mResults.classList.add("hidden");

		// Required by the BirdNET model's license (CC BY-NC-SA 4.0).
		const credit = document.createElement("p");
		credit.className = "identify-credit";
		credit.innerHTML = App.translations.AnalyzeCredit + " <a href='https://birdnet.cornell.edu' target='_blank' rel='noopener'>BirdNET</a> (K. Lisa Yang Center for Conservation Bioacoustics, Cornell Lab of Ornithology, Chemnitz University of Technology). <a href='https://creativecommons.org/licenses/by-nc-sa/4.0/' target='_blank' rel='noopener'>CC BY-NC-SA 4.0</a>";
		this.mSheet.section().appendChild(credit);

		this.mSheet.open(function () {
			self.mClosed = true;
			self.release();
		});
	}

	private start(): void {
		const self = this;
		this.mResults.classList.add("hidden");
		this.mStatus.classList.remove("error");

		// Birds are not voices: no echo cancelling, noise suppression or gain control.
		navigator.mediaDevices.getUserMedia({ audio: { echoCancellation: false, noiseSuppression: false, autoGainControl: false } }).then(function (stream) {
			if (self.mClosed) {
				stream.getTracks().forEach(t => t.stop());
				return;
			}
			self.mStream = stream;
			const AudioContextClass = window.AudioContext || (window as any).webkitAudioContext;
			const context: AudioContext = new AudioContextClass();
			self.mContext = context;
			const source = context.createMediaStreamSource(stream);
			// ScriptProcessorNode is deprecated but available everywhere, and needs no worklet file.
			const processor = context.createScriptProcessor(4096, 1, 1);
			self.mSamples = [];
			processor.onaudioprocess = function (e) {
				if (self.mRecording) {
					self.mSamples.push(new Float32Array(e.inputBuffer.getChannelData(0)));
				}
			};
			source.connect(processor);
			processor.connect(context.destination);

			self.mRecording = true;
			self.mButton.classList.add("recording");
			self.mButton.setAttribute("aria-label", App.translations.TapToStop);
			self.mButton.querySelector(".material-symbols-outlined")!.textContent = "stop";
			self.mStatus.textContent = App.translations.Recording + "... " + App.translations.TapToStop;
			self.mTimer = window.setTimeout(function () {
				self.stop();
			}, AnalyzePopUp.SECONDS * 1000);
		}).catch(function () {
			self.showError(App.translations.NoMicrophone);
		});
	}

	private stop(): void {
		if (!this.mRecording) {
			return;
		}
		this.mRecording = false;
		if (this.mTimer !== null) {
			clearTimeout(this.mTimer);
			this.mTimer = null;
		}
		const sampleRate = this.mContext ? this.mContext.sampleRate : 48000;
		const wav = AnalyzePopUp.encodeWav(this.mSamples, sampleRate);
		this.mSamples = [];
		this.release();

		this.mButton.classList.remove("recording");
		this.mButton.classList.add("busy-identify");
		this.mButton.disabled = true;
		this.mButton.querySelector(".material-symbols-outlined")!.textContent = "graphic_eq";
		this.mStatus.textContent = App.translations.IdentifyingPleaseWait;
		this.upload(wav);
	}

	// Stops the microphone and the audio processing.
	private release(): void {
		if (this.mTimer !== null) {
			clearTimeout(this.mTimer);
			this.mTimer = null;
		}
		this.mRecording = false;
		if (this.mStream) {
			this.mStream.getTracks().forEach(t => t.stop());
			this.mStream = null;
		}
		if (this.mContext) {
			this.mContext.close();
			this.mContext = null;
		}
	}

	private reset(): void {
		this.mButton.classList.remove("busy-identify", "recording");
		this.mButton.disabled = false;
		this.mButton.setAttribute("aria-label", App.translations.AnalyzeStartRecording);
		this.mButton.querySelector(".material-symbols-outlined")!.textContent = "mic";
	}

	private showError(message: string): void {
		this.release();
		this.reset();
		this.mStatus.textContent = message;
		this.mStatus.classList.add("error");
	}

	private upload(wav: Blob): void {
		const self = this;
		const form = new FormData();
		form.append("audio", wav, "recording.wav");
		// The week of the year (BirdNET's 48-week calendar) narrows the result to birds likely now.
		const now = new Date();
		const week = Math.min(48, Math.floor((now.getMonth() * 4) + Math.min(3, Math.floor((now.getDate() - 1) / 7.75))) + 1);
		form.append("meta", JSON.stringify({ save: false, num_results: 5, sensitivity: 1, week: week }));

		fetch("/api/v1/analyze", {
			method: "POST",
			body: form,
			headers: { "X-Requested-With": "XMLHttpRequest" },
			credentials: "same-origin",
		}).then(function (response) {
			if (!response.ok) {
				throw response.status;
			}
			return response.json();
		}).then(function (data) {
			if (self.mClosed) {
				return;
			}
			const birds = (data.results || []).map(e => ({ name: e[0], probability: e[1] }));
			new HttpRequest("PUT", "/v1/analyze/resolve", { birds: birds }).process(function (status: any, resolved: any) {
				self.reset();
				self.showResults(resolved.results || []);
			}, function () {
				self.showError(App.translations.SomethingWentWrong);
			});
		}).catch(function (status) {
			if (self.mClosed) {
				return;
			}
			self.showError(status === 503 || status === 502 ? App.translations.AnalyzeUnavailable : App.translations.SomethingWentWrong);
		});
	}

	private showResults(results: Array<{ species: Models.Species, probability: number }>): void {
		const self = this;
		const list = this.mResults;
		list.querySelectorAll(".identify-result, p.details").forEach(e => e.remove());
		list.classList.remove("hidden");

		if (results.length === 0) {
			this.mStatus.textContent = App.translations.NoIdentifySpecies;
			const none = document.createElement("p");
			none.className = "details";
			none.textContent = App.translations.NoIdentifyHint;
			list.appendChild(none);
			return;
		}

		this.mStatus.textContent = App.translations.RecordAgain;
		results.forEach(r => {
			const row = document.createElement("button");
			row.className = "identify-result";

			const names = document.createElement("span");
			names.className = "identify-names";
			const name = document.createElement("span");
			name.className = "identify-name";
			name.textContent = App.getTranslation(r.species.name);
			const latin = document.createElement("span");
			latin.className = "identify-latin";
			latin.textContent = r.species.name.la;
			names.appendChild(name);
			names.appendChild(latin);
			row.appendChild(names);

			const percent = Math.round(r.probability * 100);
			const confidence = document.createElement("span");
			confidence.className = "identify-confidence";
			confidence.innerHTML = "<span class='identify-meter'><span style='width: " + percent + "%'></span></span>" + percent + "%";
			row.appendChild(confidence);

			row.addEventListener("click", function () {
				self.mSheet.close();
				Ui.PopUp.whenHistorySettled(function () {
					self.mPick(r.species);
				});
			});
			list.appendChild(row);
		});
	}

	// 16-bit mono PCM WAV.
	private static encodeWav(chunks: Array<Float32Array>, sampleRate: number): Blob {
		const length = chunks.reduce((n, c) => n + c.length, 0);
		const buffer = new ArrayBuffer(44 + length * 2);
		const view = new DataView(buffer);
		const text = function (offset: number, s: string) {
			for (let i = 0; i < s.length; ++i) {
				view.setUint8(offset + i, s.charCodeAt(i));
			}
		};
		text(0, "RIFF");
		view.setUint32(4, 36 + length * 2, true);
		text(8, "WAVE");
		text(12, "fmt ");
		view.setUint32(16, 16, true);
		view.setUint16(20, 1, true);
		view.setUint16(22, 1, true);
		view.setUint32(24, sampleRate, true);
		view.setUint32(28, sampleRate * 2, true);
		view.setUint16(32, 2, true);
		view.setUint16(34, 16, true);
		text(36, "data");
		view.setUint32(40, length * 2, true);
		let offset = 44;
		chunks.forEach(chunk => {
			for (let i = 0; i < chunk.length; ++i, offset += 2) {
				const s = Math.max(-1, Math.min(1, chunk[i]));
				view.setInt16(offset, s < 0 ? s * 0x8000 : s * 0x7fff, true);
			}
		});
		return new Blob([buffer], { type: "audio/wav" });
	}
}

function openIdentifyBySound(pick: (species: Models.Species) => void): void {
	new AnalyzePopUp(pick);
}
