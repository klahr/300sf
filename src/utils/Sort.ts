namespace Utils {
    export function sortByNameOnCurrentLanguage<T extends { name: Models.Translation }>(a: T, b: T): number {
        switch (App.getLanguage()) {
            case "sv":
                return a.name.sv.localeCompare(b.name.sv);
            case "la":
                return a.name.la.localeCompare(b.name.la);
            default:
                return a.name.en.localeCompare(b.name.en);
        }
    }
}
