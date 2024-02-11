/// <reference path="ui/Div.ts"/>

class AnalyzePopUp {
	private mResultView: HTMLDivElement;
	private mAnalyzeText: HTMLParagraphElement;

	constructor(closeCallback: () => void) {
		const self = this;
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = "Identify species";
		title.style.margin = "unset";
		content.container.appendChild(title);

		const b = document.createElement("p");
		b.innerHTML = "Använd inspelningsknappen nedan för att spela in ett fågelläte för artbestämning. Det finns ingen garanti för att det stämmer, utan ska ses som en ledtråd.";
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
		recordText.innerHTML = "Click to record";
		d.appendChild(recordText);

		this.mAnalyzeText = document.createElement("p");
		this.mAnalyzeText.innerHTML = "Analyserar. Vänta...";
		this.mAnalyzeText.style.display = "none";
		d.appendChild(this.mAnalyzeText);

		const audioPlayer = document.createElement("audio");
		audioPlayer.setAttribute("controls", "controls");
		//content.container.appendChild(audioPlayer);

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
						recordText.innerHTML = "Click to record.";
					}, 10000);
					recordButton.innerHTML = "stop_circle";
					recordText.innerHTML = "Recording...";
				} else {
					clearTimeout(timeoutId);
					mediaRecorder.stop();
					recordButton.innerHTML = "album";
					recordText.innerHTML = "Click to record.";
				}
			});
		};

		const handleError = (error) => {
			console.error('Error accessing microphone:', error);
		};

		navigator.mediaDevices.getUserMedia({ audio: true })
			.then(handleSuccess)
			.catch(handleError);

		new Ui.PopUp(content.container, function () {
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
			console.log('Server response:', data);

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
						const r = document.createElement("p");
						r.innerHTML = App.getTranslation(e.species.name);
						self.mResultView.appendChild(r);
					
						const funding = new Ui.ProgressBar();
						funding.style.position = "relative";
						funding.setProgress(e.probability * 100);
						funding.style.width = "100%";
						funding.style.height = "8px";
						self.mResultView.appendChild(funding.container);
					});
				} else {
					const r = document.createElement("p");
					r.innerHTML = "Lyckades inte identifiera arten.";
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

for (let i = 0; i < 5; ++i) {

}