/// <reference path="Order.ts"/>
/// <reference path="Translation.ts"/>

namespace Models {
	export class Species {
		uuid: string;
		name: Translation;
		order: Order;
	}
}
