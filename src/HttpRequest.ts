/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>
/// <reference path="AppConfig.ts"/>

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
			this.setHeader("Authorization", "Bearer " + accessToken);
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
				} else if (xhr.status == 401) {
					const ks = new LocalStorage();
					if (xhr.responseJSON.code === 7001) {
						ks.clear();
						window.open("/", "_self");
						return;
					}
					const d = {
						grant_type: "refresh_token",
						refresh_token: ks.get("x300sf_rt")
					}
					const r = new HttpRequest("POST", "/v1/auth", d);
					r.process(function (status: any, data: Responses.AuthenticationResponse) {
						ks.set("x300sf_at", data.access_token, data.expires_in);
						ks.set("x300sf_rt", data.refresh_token, 3600 * 7 * 60);
						ks.set("lang", data.lang, data.expires_in);
						ks.set("scope", data.scope, data.expires_in);
						self.setHeader("Authorization", "Bearer " + data.access_token);
						self.process(success, fail, complete);
					}, function (status: any, data: Responses.GeneralErrorResponse, err: any) {
						ks.clear();
						window.open("/", "_self");
					});
				} else {
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
