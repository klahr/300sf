/// <reference path="ui/Progress.ts"/>
/// <reference path="App.ts"/>
/// <reference path="responses/Error.ts"/>
/// <reference path="AppConfig.ts"/>
/// <reference path="Session.ts"/>

declare let $: any;

class HttpRequest {
	private mHeaders = new Map<string, string>();
	private mAsync = true;
	private mNoRetry = false;
	private mRetried = false;
	private mDataType = "json";
	private mQuiet = false;
	private mTimeLimit = 60000;
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

	// For the session's own requests: a 401 is reported, never refreshed.
	// For responses that are not JSON, such as the CSV export.
	public setTextResponse(): void {
		this.mDataType = "text";
	}

	public setNoRetry(): void {
		this.mNoRetry = true;
	}

	// No connection, an overloaded server or a timeout go to the fail handler, with status 0, 429
	// or 503, instead of a dialog: for requests that know what to do offline.
	public setQuiet(): void {
		this.mQuiet = true;
	}

	public setTimeLimit(ms: number): void {
		this.mTimeLimit = ms;
	}

	public process(success: (status: any, data: any) => void = function () { }, fail: (status: any, data: Responses.GeneralErrorResponse, error: any) => void = function () { }, complete: () => void = function () { }): void {
		const self = this;

		// The session cookie authenticates the request; this header proves it
		// comes from the app and not from another site.
		self.setHeader("X-Requested-With", "XMLHttpRequest");

		$.ajax({
			beforeSend: function (xhr) {
				for (const [key, value] of self.mHeaders.entries()) {
					xhr.setRequestHeader(key, value);
				}
				++HttpRequest.sNumRequests;
				Ui.Progress.start();
			},
			type: self.mMethod,
			data: JSON.stringify(self.mData),
			async: self.mAsync,
			url: AppConfig.ApiEndpoint + self.mUrl,
			success: function (data, _textStatus, xhr) {
				// A failing handler must not skip the clean-up below, such as hiding the progress bar.
				try {
					success(xhr.status, data);
				} catch (e) {
					console.error(e);
				}
			},
			error: function (xhr, _textStatus, errorThrown) {
				if ((xhr.status === 503 || xhr.status === 429 || xhr.status === 0) && self.mQuiet) {
					fail(xhr.status, xhr.responseJSON, errorThrown);
				} else if (xhr.status === 503 || xhr.status === 429 || xhr.status === 0) {
					new Ui.MessageBox(App.translations.ConnectionError, App.translations.ConnectionErrorDescription, null, "var(--error-color)", () => void {});
				} else if (xhr.status == 401 && !self.mNoRetry && xhr.responseJSON?.code !== Responses.ErrorCode.ErrorMissingPermission) {
					// Expired or missing access token: refresh once and retry once.
					if (self.mRetried || !Session.exists) {
						if (Session.exists) {
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
			timeout: self.mTimeLimit,
			dataType: self.mDataType,
			contentType: "application/json",
			complete: function () {
				--HttpRequest.sNumRequests;
				Ui.Progress.done();
				complete();
			},
		});
	}
}
