/// <reference path="ui/Div.ts"/>
/// <reference path="App.ts"/>

class AnalyzePopUp {
	private mResultView: HTMLDivElement;
	private mAnalyzeText: HTMLParagraphElement;
	private mPopUp: Ui.PopUp;

	constructor(closeCallback: () => void) {
		const self = this;
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = App.translations.IdentifySpecies;
		title.style.margin = "unset";
		content.container.appendChild(title);

		const b = document.createElement("p");
		b.innerHTML = App.translations.AnalyzeDescription;
		content.container.appendChild(b);

		const d = document.createElement("div");
		d.style.textAlign = "center";
		d.style.width = "100%";
		d.style.margin = "32px 0";
		content.container.appendChild(d);

		const recordButton = document.createElement("span");
		recordButton.className = "material-symbols-outlined";
		recordButton.style.cursor = "pointer";
		recordButton.style.fontSize = "128px";
		recordButton.style.color = "var(--error-color)";
		recordButton.innerHTML = "album";
		recordButton.style.backgroundColor = "white";
		recordButton.style.borderRadius = "50%";
		d.appendChild(recordButton);

		const recordText = document.createElement("p");
		recordText.innerHTML = App.translations.AnalyzeStartRecording;
		d.appendChild(recordText);

		this.mAnalyzeText = document.createElement("p");
		this.mAnalyzeText.innerHTML = App.translations.IdentifyingPleaseWait;
		this.mAnalyzeText.style.display = "none";
		d.appendChild(this.mAnalyzeText);

		const audioPlayer = document.createElement("audio");
		audioPlayer.setAttribute("controls", "controls");
		//content.container.appendChild(audioPlayer);

		const resultHeader = document.createElement("h2");
		resultHeader.innerHTML = App.translations.Result;
		resultHeader.style.margin = "unset";
		content.container.appendChild(resultHeader);

		this.mResultView = document.createElement("div");
		content.container.appendChild(this.mResultView);

		let mediaRecorder;
		let audioChunks = [];

		const handleSuccess = (stream) => {
			mediaRecorder = new MediaRecorder(stream);
			mediaRecorder.ondataavailable = (event) => {
				if (event.data.size > 0) {
					audioChunks.push(event.data);
				}
			};

			mediaRecorder.onstop = () => {
				const audioBlob = new Blob(audioChunks, { type: 'audio/wav' });
				audioChunks = [];
		
				const audioUrl = URL.createObjectURL(audioBlob);
				audioPlayer.src = audioUrl;

				self.uploadToServer(audioBlob);
			};

			let timeoutId;
			recordButton.addEventListener('click', () => {
				if (mediaRecorder.state === 'inactive') {
					audioChunks = [];
					mediaRecorder.start();
					timeoutId = setTimeout(function() {
						mediaRecorder.stop();
						recordButton.innerHTML = "album";
						recordText.innerHTML = App.translations.AnalyzeStartRecording;
					}, 10000);
					recordButton.innerHTML = "stop_circle";
					recordText.innerHTML = App.translations.Recording + "...";
				} else {
					clearTimeout(timeoutId);
					mediaRecorder.stop();
					recordButton.innerHTML = "album";
					recordText.innerHTML = App.translations.AnalyzeStartRecording;
				}
			});
		};

		const handleError = (error) => {
			console.error('Error accessing microphone:', error);
			new Ui.MessageBox(App.translations.Microphone, App.translations.NoMicrophone, null, "var(--error-color)", function () {
				self.mPopUp.close();
			});
		};

		navigator.mediaDevices.getUserMedia({ audio: true })
			.then(handleSuccess)
			.catch(handleError);

		this.mPopUp = new Ui.PopUp(content.container, function () {
			closeCallback();
		});
	}

	private uploadToServer(audioBlob: Blob) {
		const self = this;
		const formData = new FormData();
		formData.append('audio', audioBlob, 'recording.wav');
		formData.append('meta', '{"save":false}');
		this.mAnalyzeText.style.display = "";

		fetch('https://api.300sf.se/analyze', {
			method: 'POST',
			body: formData,
		})
		.then(response => response.json())
		.then(data => {
			let content: { name: string; probability: number }[] = [];

			self.mResultView.innerHTML = "";
			data.results.forEach(e => {
				content.push({
					"name": e[0],
					"probability": e[1]
				});
			});

			const r = new HttpRequest("PUT", "/v1/analyze/resolve", {"birds": content});
			r.process((status: any, data: any) => {
				if (data.results) {
					data.results.forEach(e => {
						const d = document.createElement("div");
						d.style.border = "2px solid #333";
						d.style.padding = "8px";
						d.style.borderRadius = "8px";
						d.style.marginBottom = "4px";
						d.style.cursor = "pointer";
						d.addEventListener("click", function() {
							openSpeciesPopUp(e.species, null, () => {});
						});
						self.mResultView.appendChild(d);
						const r = document.createElement("p");
						r.innerHTML = App.getTranslation(e.species.name);
						d.appendChild(r);
					
						const probability = new Ui.ProgressBar();
						probability.style.position = "relative";
						probability.setProgress(e.probability * 100);
						probability.style.width = "100%";
						probability.style.height = "8px";
						d.appendChild(probability.container);
					});
				} else {
					const r = document.createElement("p");
					r.innerHTML = App.translations.NoIdentifySpecies;
					self.mResultView.appendChild(r);
				}
				this.mAnalyzeText.style.display = "none";
			});
		})
		.catch(error => {
			console.error('Error uploading to server:', error);
		});
	}
}
