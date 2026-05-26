export const THEMES = [
	{
		// Midnight Indigo/Purple
		cover: "radial-gradient(circle at top left, hsl(262, 70%, 36%) 0%, hsl(262, 80%, 12%) 100%)",
		avatarText: "hsl(262, 80%, 46%)",
		avatarBg: "linear-gradient(135deg, hsl(262, 70%, 96%) 0%, hsl(262, 70%, 90%) 100%)"
	},
	{
		// Teal/Emerald (Nature/Sustainability)
		cover: "radial-gradient(circle at top left, hsl(162, 75%, 28%) 0%, hsl(162, 85%, 8%) 100%)",
		avatarText: "hsl(162, 80%, 30%)",
		avatarBg: "linear-gradient(135deg, hsl(162, 70%, 94%) 0%, hsl(162, 70%, 86%) 100%)"
	},
	{
		// Sunset Amber/Rose
		cover: "radial-gradient(circle at top left, hsl(340, 75%, 38%) 0%, hsl(340, 85%, 12%) 100%)",
		avatarText: "hsl(340, 80%, 45%)",
		avatarBg: "linear-gradient(135deg, hsl(340, 70%, 96%) 0%, hsl(340, 70%, 90%) 100%)"
	},
	{
		// Ocean Sapphire/Blue
		cover: "radial-gradient(circle at top left, hsl(210, 80%, 35%) 0%, hsl(210, 90%, 12%) 100%)",
		avatarText: "hsl(210, 80%, 45%)",
		avatarBg: "linear-gradient(135deg, hsl(210, 70%, 95%) 0%, hsl(210, 70%, 88%) 100%)"
	},
	{
		// Gold/Bronze (Premium/Prestige)
		cover: "radial-gradient(circle at top left, hsl(36, 80%, 36%) 0%, hsl(36, 90%, 12%) 100%)",
		avatarText: "hsl(36, 80%, 35%)",
		avatarBg: "linear-gradient(135deg, hsl(36, 70%, 94%) 0%, hsl(36, 70%, 86%) 100%)"
	}
];

export function getTheme(name = "") {
	if (!name) return THEMES[0];
	let sum = 0;
	for (let i = 0; i < name.length; i++) {
		sum += name.charCodeAt(i);
	}
	const index = sum % THEMES.length;
	return THEMES[index];
}

export function getInitials(name) {
	if (!name) return "O";
	const words = name.replace(/[^\p{L}\p{N}\s]/gu, "").trim().split(/\s+/).filter(Boolean);
	if (words.length === 0 || !words[0]) return "O";
	
	if (words.length === 1) {
		return Array.from(words[0]).slice(0, 2).join("").toLocaleUpperCase();
	}
	
	return `${Array.from(words[0])[0]}${Array.from(words[1])[0]}`.toLocaleUpperCase();
}
