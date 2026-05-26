const SDG_COLORS = {
	1: "#e5243b",
	2: "#dda63a",
	3: "#4c9f38",
	4: "#c5192d",
	5: "#ff3a21",
	6: "#26bde2",
	7: "#fcc30b",
	8: "#a21942",
	9: "#fd6925",
	10: "#dd1367",
	11: "#fd9d24",
	12: "#bf8b2e",
	13: "#3f7e44",
	14: "#0a97d9",
	15: "#56c02b",
	16: "#00689d",
	17: "#19486a",
};

const SDG_TEXT_COLORS = {
	6: "#111111",
	7: "#111111",
	11: "#111111",
	15: "#111111",
};

const SDG_TITLES = {
	id: {
		1: "Tanpa Kemiskinan",
		2: "Tanpa Kelaparan",
		3: "Kehidupan Sehat dan Sejahtera",
		4: "Pendidikan Berkualitas",
		5: "Kesetaraan Gender",
		6: "Air Bersih dan Sanitasi Layak",
		7: "Energi Bersih dan Terjangkau",
		8: "Pekerjaan Layak dan Pertumbuhan Ekonomi",
		9: "Industri, Inovasi dan Infrastruktur",
		10: "Berkurangnya Kesenjangan",
		11: "Kota dan Permukiman Berkelanjutan",
		12: "Konsumsi dan Produksi yang Bertanggung Jawab",
		13: "Penanganan Perubahan Iklim",
		14: "Ekosistem Lautan",
		15: "Ekosistem Daratan",
		16: "Perdamaian, Keadilan dan Kelembagaan yang Tangguh",
		17: "Kemitraan untuk Mencapai Tujuan",
	},
	en: {
		1: "No Poverty",
		2: "Zero Hunger",
		3: "Good Health and Well-being",
		4: "Quality Education",
		5: "Gender Equality",
		6: "Clean Water and Sanitation",
		7: "Affordable and Clean Energy",
		8: "Decent Work and Economic Growth",
		9: "Industry, Innovation and Infrastructure",
		10: "Reduced Inequalities",
		11: "Sustainable Cities and Communities",
		12: "Responsible Consumption and Production",
		13: "Climate Action",
		14: "Life Below Water",
		15: "Life on Land",
		16: "Peace, Justice and Strong Institutions",
		17: "Partnerships for the Goals",
	},
};

export const SDG_NUMBERS = Array.from({ length: 17 }, (_, index) => index + 1);

function iconPath(number) {
	return number ? `/sdgs/E-WEB-Goal-${`${number}`.padStart(2, "0")}.png` : "";
}

function readGoalText(goal) {
	if (goal == null) {
		return "";
	}

	if (typeof goal === "number") {
		return `${goal}`;
	}

	if (typeof goal === "string") {
		return goal;
	}

	if (typeof goal === "object") {
		return `${goal.code || goal.sdg_code || goal.goal || goal.number || goal.title || goal.name || ""}`;
	}

	return "";
}

function parseGoalNumber(goal) {
	const text = readGoalText(goal);
	const match = text.match(/\b(?:sdg|sdgs|goal)?\s*0?([1-9]|1[0-7])\b/i);
	const number = match ? Number(match[1]) : Number.NaN;
	return Number.isInteger(number) && number >= 1 && number <= 17 ? number : null;
}

function readCustomTitle(goal) {
	if (!goal || typeof goal !== "object" || Array.isArray(goal)) {
		return "";
	}

	return `${goal.title || goal.name || goal.label || ""}`.trim();
}

export function normalizeSdgGoals(source, language = "id") {
	const items =
		Array.isArray(source) ? source
		: typeof source === "string" ? source.split(/[\n,]/)
		: [];
	const seen = new Set();

	return items
		.map((item) => {
			const raw = readGoalText(item).trim();
			const number = parseGoalNumber(item);
			const key = number || raw.toLowerCase();

			if (!key || seen.has(key)) {
				return null;
			}

			seen.add(key);
			const titles = SDG_TITLES[language] || SDG_TITLES.id;

			return {
				id: number ? `sdg-${number}` : raw,
				number,
				code: number ? `${number}`.padStart(2, "0") : "SDG",
				title: readCustomTitle(item) || (number ? titles[number] : raw),
				raw,
				color: number ? SDG_COLORS[number] : "var(--purple-700)",
				textColor: number ? SDG_TEXT_COLORS[number] || "#ffffff" : "#ffffff",
				icon: iconPath(number),
			};
		})
		.filter(Boolean);
}

export function getSdgGoal(number, language = "id") {
	return normalizeSdgGoals([number], language)[0];
}

export function getAllSdgGoals(language = "id") {
	return SDG_NUMBERS.map((number) => getSdgGoal(number, language));
}
