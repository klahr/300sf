/// <reference path="ui/Div.ts"/>

class AnalyzePopUp {
	constructor(closeCallback: () => void) {
		const self = this;
		const content = new Ui.Div();
		content.style.textAlign = "left";
		content.style.width = "100%";
		content.style.position = "relative";

		const title = document.createElement("h1");
		title.innerHTML = "Analyze";
		title.style.margin = "unset";
		content.container.appendChild(title);

		const recordButton = document.createElement("button");
		recordButton.innerHTML = "Record";
		content.container.appendChild(recordButton);

		const uploadButton = document.createElement("button");
		uploadButton.innerHTML = "Upload";
		content.container.appendChild(uploadButton);

		const audioPlayer = document.createElement("audio");
		audioPlayer.setAttribute("controls", "controls");
		content.container.appendChild(audioPlayer);

		const resultView = document.createElement("div");
		content.container.appendChild(resultView);

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

				uploadButton.style.display = 'inline'; // Show the upload button
				uploadButton.addEventListener('click', () => {
				  uploadToServer(audioBlob);
				});
			};

			recordButton.addEventListener('click', () => {
				if (mediaRecorder.state === 'inactive') {
					audioChunks = [];
					mediaRecorder.start();
					setTimeout(() => mediaRecorder.stop(), 10000); // Record for 10 seconds
					recordButton.textContent = 'Stop';
				} else {
					mediaRecorder.stop();
					recordButton.textContent = 'Record';
				}
			});
		};

		const handleError = (error) => {
			console.error('Error accessing microphone:', error);
		};

		const uploadToServer = (audioBlob) => {
			const formData = new FormData();
			formData.append('audio', audioBlob, 'recording.wav');
			formData.append('meta', '{"save":false}');

			fetch('https://api.300sf.se/analyze', {
				method: 'POST',
				body: formData,
			})
			.then(response => response.json())
			.then(data => {
				console.log('Server response:', data);

				let content: { name: string; probability: number }[] = [];

				resultView.innerHTML = "";
				data.results.forEach(e => {
					content.push({
						"name": e[0],
						"probability": e[1]
					});
				});

				const r = new HttpRequest("PUT", "/v1/analyze/resolve", {"birds": content});
				r.process((status: any, data: any) => {
					console.log(data);
					if (data.results) {
						data.results.forEach(e => {
							const r = document.createElement("p");
							r.innerHTML = App.getTranslation(e.species.name) + " (" + Math.round(e.probability * 100) + "%)";
							resultView.appendChild(r);
						});
					} else {
						const r = document.createElement("p");
						r.innerHTML = "No result.";
						resultView.appendChild(r);
					}
				});
			})
			.catch(error => {
				console.error('Error uploading to server:', error);
			});
		};

		navigator.mediaDevices.getUserMedia({ audio: true })
			.then(handleSuccess)
			.catch(handleError);

		new Ui.PopUp(content.container, function () {
			closeCallback();
		});
	}
}