/// <reference path="Cookie.ts"/>
/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>


declare let $: any;

class HttpRequest {
	private mHeaders = new Map<string, string>();
	private mAsync = true;

	constructor(private mMethod: string, private mUrl: string, private mData: any = null) {
		this.setHeader("Accept", "application/json");
		this.setHeader("Content-Type", "application/json");
		this.setHeader("X-App-Language", App.language);
		const accessToken = getCookie("x300sf_at");
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

	public process(success: (status: any, data: any) => void = function() {}, fail: (status: any, data: Responses.Error, error: any) => void = function() {}, complete: () => void = function() {}): void {
		const self = this;
		$.ajax({
			beforeSend: function(xhr) {
				for (const [key, value] of self.mHeaders.entries()) {
					xhr.setRequestHeader(key, value);
				}
			},
			type: self.mMethod,
			data: JSON.stringify(self.mData),
			async: self.mAsync,
			url: Config.ApiEndpoint + self.mUrl,
			success: function(data, _textStatus, xhr) {
				const cl = xhr.getResponseHeader('Content-Language');
				if (cl) {
					setCookie("lang", cl);
					App.setLanguage(cl);
				}
				success(xhr.status, data);
			},
			error: function(xhr, _textStatus, errorThrown) {
				const response = xhr.responseJSON as Responses.Error;
				switch (response.code) {
					case Responses.ErrorCode.ErrorLoggedOut:
					default:
						if (response.code === Responses.ErrorCode.ErrorLoggedOut) {
							// TODO: Refresh session instead of accept logged out.
							window.open("/?redirect=" + window.location.href, "_self");
						}
					break;
				}
				fail(xhr.status, xhr.responseJSON, errorThrown);
			},
			timeout: 20000,
			dataType: "json",
			contentType: "application/json",
			complete: function() {
				complete();
			},
		});
	}
}
