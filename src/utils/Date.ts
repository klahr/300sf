namespace Utils {
	export function isValidDate(date: Date): boolean {
		if (date.getTime() === -62135596800000) {
			return false;
		}
		return true;
	}

	export function toYYYY_MM_DD_hh_mm(date: Date): string {
		const d = new Date(date);
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0'); // Month is 0-based, so we add 1 and format it with leading zero.
		const day = String(d.getDate()).padStart(2, '0');
		const hours = String(d.getHours()).padStart(2, '0');
		const minutes = String(d.getMinutes()).padStart(2, '0');
		return `${year}-${month}-${day} ${hours}:${minutes}`;
	}

	export function toYYYY_MM_DD(date: Date): string {
		const d = new Date(date);
		const year = d.getFullYear();
		const month = String(d.getMonth() + 1).padStart(2, '0'); // Month is 0-based, so we add 1 and format it with leading zero.
		const day = String(d.getDate()).padStart(2, '0');
		return `${year}-${month}-${day}`;
	}

	export function toISODate(date: Date): string {
		var r = new Date(date.getTime()).toISOString();
		return r.slice(0, r.lastIndexOf(".")) + 'Z';
	}

	export function getTimezoneOffset(): number {
		const now = new Date();
		const isDaylightSavingTime = now.getTimezoneOffset() !== new Date(now.getFullYear(), 0, 1).getTimezoneOffset();
		return -now.getTimezoneOffset() + (isDaylightSavingTime ? 60 : 0);
	}
}
