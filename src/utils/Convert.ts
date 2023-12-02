namespace Utils {
	export function toInt(str: string): number {
		if (str === "") {
			return 0;
		}
		return parseInt(str);
	}
}