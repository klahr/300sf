/// <reference path="utils/Date.ts"/>
/// <reference path="Page.ts"/>
/// <reference path="ui/Input.ts"/>

class ContactPage extends Page {
	private mSubject: Ui.Input;
	private mMessage: HTMLTextAreaElement;
	private mSendButton: HTMLButtonElement;

	constructor(appListener: IAppListener) {
		super(appListener);

		const self = this;

		this.addProfile();

		const title = document.createElement("h1");
		title.innerHTML = App.translations.Contact;
		this.container.appendChild(title);

		const contactDescription = document.createElement("p");
		contactDescription.innerHTML = App.translations.ContactDescription;
		this.container.appendChild(contactDescription);

		this.mSubject = new Ui.Input("text");
		this.mSubject.setPlaceHolder(App.translations.Subject);
		this.mSubject.container.addEventListener("input", function() {
			self.verify();
		});
		this.attach(this.mSubject);

		this.mMessage = document.createElement("textarea");
		this.mMessage.style.height = "128px";
		this.mMessage.placeholder = App.translations.Message;
		this.mMessage.addEventListener("input", function() {
			self.verify();
		});
		this.container.appendChild(this.mMessage);	

		this.mSendButton = document.createElement("button");
		this.mSendButton.innerHTML = App.translations.Send;
		this.mSendButton.disabled = true;
		this.mSendButton.addEventListener("click", function() {
			const d = {
				subject: self.mSubject.value,
				message: self.mMessage.value
			};

			self.mSendButton.disabled = true;
			self.mSendButton.innerHTML = App.translations.Sending;
			self.mSubject.container.setAttribute("disabled", "disabled");
			self.mMessage.setAttribute("disabled", "disabled");
			const r = new HttpRequest("POST", "/v1/users/sdmigrate", d);
			r.process(function(status: any, data: any) {
				new Ui.MessageBox("Success", "Message sent!", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--success-color)", function() {});
				self.mMessage.value = "";
				self.mSubject.value = "";
				self.refresh();
			}, function() {
				new Ui.MessageBox("Oh, no!", "Message could not be delivered.", "/img/3f638133-a433-439b-919c-fdef170b882e.png", "var(--error-color)", function() {});
			}, function() {
				self.mSendButton.innerHTML = App.translations.Send;
				self.mSendButton.disabled = false;
				self.mSubject.container.removeAttribute("disabled");
				self.mMessage.removeAttribute("disabled");
			});
		});

		this.container.appendChild(this.mSendButton);
	}

	private verify(): void {
		this.mSendButton.disabled = this.mSubject.value.length < 4 || this.mMessage.value.length < 16;
	}

	public get id(): string {
		return "contact";
	}

	public onBecomeVisible(): void {
		super.onBecomeVisible();
		this.refresh();
	}

	private refresh(): void {
		if (App.user) {
			this.mProfile.setImage("/img/portraits/" + App.user.portrait);
		}
	}
}