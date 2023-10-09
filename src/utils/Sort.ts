/// <reference path="../App.ts"/>

namespace Utils {
    export function sortByNameOnCurrentLanguage<T extends { name: Models.Translation }>(a: T, b: T): number {
        switch (App.language) {
            case "sv":
                return a.name.sv.localeCompare(b.name.sv);
            case "la":
                return a.name.la.localeCompare(b.name.la);
            case "de":
                return a.name.de.localeCompare(b.name.de);
            default:
                return a.name.en.localeCompare(b.name.en);
        }
    }
}