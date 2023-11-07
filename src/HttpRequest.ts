/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>
/// <reference path="AppConfig.ts"/>
/// <reference path="SignInBox.ts"/>

declare let $: any;

class HttpRequest {
	private mHeaders = new Map<string, string>();
	private mAsync = true;

	constructor(private mMethod: string, private mUrl: string, private mData: any = {}) {
		this.setHeader("Accept", "application/json");
		this.setHeader("Content-Type", "application/json");
		this.setHeader("X-App-Language", App.language);
		const ks = new LocalStorage();
		const accessToken = ks.get("x300sf_at");
		if (accessToken != null && accessToken != "") {
			this.setHeader("X-Access-Token", accessToken);
		}
	}

	public setHeader(key: string, value: string): void {
		this.mHeaders.set(key, value);
	}

	public setSync(): void {
		this.mAsync = false;
	}

	public process(success: (status: any, data: any) => void = function () { }, fail: (status: any, data: Responses.GeneralErrorResponse, error: any) => void = function () { }, complete: () => void = function () { }): void {
		const self = this;

		$.ajax({
			beforeSend: function (xhr) {
				for (const [key, value] of self.mHeaders.entries()) {
					xhr.setRequestHeader(key, value);
				}
			},
			type: self.mMethod,
			data: JSON.stringify(self.mData),
			async: self.mAsync,
			url: AppConfig.ApiEndpoint + self.mUrl,
			success: function (data, _textStatus, xhr) {
				success(xhr.status, data);
			},
			error: function (xhr, _textStatus, errorThrown) {
				if (xhr.status === 503 || xhr.status === 0) {
					new Ui.MessageBox(App.translations.ConnectionError, App.translations.ConnectionErrorDescription, null, "var(--error-color)", () => void {});
				} else {
					const response = xhr.responseJSON as Responses.GeneralErrorResponse;
					if (response) {
						switch (response.code) {
							case Responses.ErrorCode.ErrorLoggedOut:
								if (response.code === Responses.ErrorCode.ErrorLoggedOut) {
									new SignInBox(App.translations.SessionExpired, App.translations.SessionExpiredDescription, null, "var(--success-color)", () => { }, () => { }, App.translations.SignIn, App.translations.Cancel);
								}
								return;
							default:
								break;
						}
					}
					fail(xhr.status, xhr.responseJSON, errorThrown);
				}
			},
			timeout: 60000,
			dataType: "json",
			contentType: "application/json",
			complete: function () {
				complete();
			},
		});
	}
}
