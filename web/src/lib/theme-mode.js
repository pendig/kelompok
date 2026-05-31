import { browser } from "$app/environment";
import { writable } from "svelte/store";

const STORAGE_KEY = "kelompok:theme";

function systemTheme() {
	return browser && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function initialTheme() {
	if (!browser) return "light";
	const saved = localStorage.getItem(STORAGE_KEY);
	return saved === "light" || saved === "dark" ? saved : systemTheme();
}

export const theme = writable(initialTheme());

function apply(value) {
	if (browser) {
		document.documentElement.dataset.theme = value;
		localStorage.setItem(STORAGE_KEY, value);
	}
}

export function toggleTheme() {
	theme.update((current) => {
		const next = current === "dark" ? "light" : "dark";
		apply(next);
		return next;
	});
}

export function initTheme() {
	if (!browser) return;
	const current = initialTheme();
	theme.set(current);
	document.documentElement.dataset.theme = current;
}
