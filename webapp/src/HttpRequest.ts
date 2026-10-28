/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>
/// <reference path="AppConfig.ts"/>
/// <reference path="Session.ts"/>

declare let $: any;

class HttpRequest {
	private mHeaders = new Map<string, string>();
	private mAsync = true;
	private mAnonymous = false;
	private mRetried = false;
	private static sNumRequests = 0;

	constructor(private mMethod: string, private mUrl: string, private mData: any = {}) {
		this.setHeader("Accept", "application/json, text/csv");
		this.setHeader("Content-Type", "application/json");
		this.setHeader("X-App-Language", App.language);
	}

	public setHeader(key: string, value: string): void {
		this.mHeaders.set(key, value);
	}

	public setSync(): void {
		this.mAsync = false;
	}

	// Sent without the access token and never triggers a token refresh.
	public setAnonymous(): void {
		this.mAnonymous = true;
	}

	public process(success: (status: any, data: any) => void = function () { }, fail: (status: any, data: Responses.GeneralErrorResponse, error: any) => void = function () { }, complete: () => void = function () { }): void {
		const self = this;

		const accessToken = self.mAnonymous ? null : Session.accessToken;
		if (accessToken) {
			self.setHeader("Authorization", "Bearer " + accessToken);
		} else {
			self.mHeaders.delete("Authorization");
		}

		$.ajax({
			beforeSend: function (xhr) {
				for (const [key, value] of self.mHeaders.entries()) {
					xhr.setRequestHeader(key, value);
				}
				++HttpRequest.sNumRequests;
				const loading = document.getElementById("loading");
				if (loading) {
					loading.style.display = "block";
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
				if (xhr.status === 503 || xhr.status === 429 || xhr.status === 0) {
					new Ui.MessageBox(App.translations.ConnectionError, App.translations.ConnectionErrorDescription, null, "var(--error-color)", () => void {});
				} else if (xhr.status == 401 && !self.mAnonymous && xhr.responseJSON?.code !== Responses.ErrorCode.ErrorMissingPermission) {
					// Expired or missing access token: refresh once and retry once.
					if (self.mRetried || Session.refreshToken === null) {
						if (Session.refreshToken !== null || accessToken) {
							Session.signInAgain();
							return;
						}
						fail(xhr.status, xhr.responseJSON, errorThrown);
						return;
					}
					Session.refresh(function (ok: boolean) {
						if (ok) {
							self.mRetried = true;
							self.process(success, fail, complete);
						} else {
							Session.signInAgain();
						}
					});
				} else {
					fail(xhr.status, xhr.responseJSON, errorThrown);
				}
			},
			timeout: 60000,
			dataType: "json",
			contentType: "application/json",
			complete: function () {
				--HttpRequest.sNumRequests;
				if (HttpRequest.sNumRequests == 0) {
					const loading = document.getElementById("loading");
					if (loading) {
						loading.style.display = "none";
					}
				}
				complete();
			},
		});
	}
}
