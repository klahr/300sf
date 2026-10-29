// The colours a list, tag, competition or announcement can have. All have the same luminance, so
// white text on them and a dot of them on either theme's paper stay readable. The first is the default.
namespace Colors {
	export const palette = [
		"#1273a3", // blue
		"#157975", // teal
		"#297b44", // green
		"#8d651e", // ochre
		"#af5019", // orange
		"#c5382e", // red
		"#c13082", // pink
		"#9546c9", // purple
		"#5262cd", // indigo
		"#606d7f", // slate
	];
	export const fallback = palette[0];
}
