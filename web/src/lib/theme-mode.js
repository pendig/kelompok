import { browser } from "$app/environment";
import { writable } from "svelte/store";

const STORAGE_KEY = "kelompok:theme";

function systemTheme() {
	return browser && window.matchMedia?.("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function savedTheme() {
	if (!browser) return null;

	try {
		const saved = localStorage.getItem(STORAGE_KEY);
		return saved === "light" || saved === "dark" ? saved : null;
	} catch (error) {
		return null;
	}
}

function initialTheme() {
	if (!browser) return "light";
	return savedTheme() ?? systemTheme();
}

export const theme = writable("light");

function apply(value) {
	if (!browser) return;

	document.documentElement.dataset.theme = value;

	try {
		localStorage.setItem(STORAGE_KEY, value);
	} catch (error) {
		// Theme still applies for this page view when storage is unavailable.
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
