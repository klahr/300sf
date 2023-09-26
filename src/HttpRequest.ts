/// <reference path="Cookie.ts"/>
/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>

declare let $: any;

class HttpRequest {
	private mHeaders = new Map<string, string>();

	constructor(private mMethod: string, private mUrl: string, private mData: any = null) {
		this.setHeader("Accept", "application/json");
		this.setHeader("Content-Type", "text/plain");
		this.setHeader("X-App-Language", App.getLanguage());
		const accessToken = getCookie("x300sf_at");
		if (accessToken != null && accessToken != "") {
			this.setHeader("X-Access-Token", accessToken);
		}
		if (App.apiKey != null) {
			this.setHeader("X-Api-Key", App.apiKey);
		}
	}

	public setHeader(key: string, value: string): void {
		this.mHeaders.set(key, value);
	}

	public process(success: (status: any, data: any) => void = function() {}, fail: (status: any, data: any, error: any) => void = function() {}, complete: () => void = function() {}): void {
		const self = this;
		$.ajax({
			beforeSend: function(xhr) {
				for (const [key, value] of self.mHeaders.entries()) {
					xhr.setRequestHeader(key, value);
				}
			},
			type: self.mMethod,
			data: self.mData,
			url: Config.ApiEndpoint + self.mUrl,
			success: function(data, _textStatus, xhr) {
				success(xhr.status, data);
			},
			error: function(xhr, _textStatus, errorThrown) {
				const response = xhr.responseJSON as Responses.Error;
				switch (response.code) {
					case Responses.ErrorCode.ErrorLoggedOut:
					default:
						// TODO: Refresh session instead of accept logged out.
						location.reload();
					break;
				}
				fail(xhr.status, xhr.responseJSON, errorThrown);
			},
			timeout: 20000,
			dataType: "json",
			complete: function() {
				complete();
			},
		});
	}
}
