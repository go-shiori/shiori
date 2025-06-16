// Handles API responses in both legacy and new message formats
export async function handleApiResponse(response) {
	if (!response.ok) throw response;

	// Return early for 204 No Content responses
	if (response.status === 204) {
		return null;
	}

	const contentType = response.headers.get("Content-Type");
	if (!contentType || !contentType.includes("application/json")) {
		return response;
	}

	const data = await response.json();

	// Check if response is in the new message format
	if (data && typeof data === "object" && "ok" in data && "message" in data) {
		if (!data.ok) {
			throw new Error(data.message?.error || "Unknown error");
		}
		return data.message;
	}

	// Legacy format - return as is
	return data;
}

// Handles API errors and returns a user-friendly error message
export async function handleApiError(error) {
	if (error instanceof Response) {
		const data = await error.json();

		if (data && typeof data === "object" && "error" in data) {
			return data.error;
		} else if (
			data &&
			typeof data === "object" &&
			"message" in data &&
			"error" in data.message
		) {
			return data.message.error;
		} else {
			return error.statusText;
		}
	}

	return "Unknown error occurred";
}

// Makes an API request with proper error handling
export async function apiRequest(url, options = {}) {
	try {
		const response = await fetch(url, {
			...options,
			headers: {
				"Content-Type": "application/json",
				Authorization: "Bearer " + localStorage.getItem("shiori-token"),
				...(options.headers || {}),
			},
		});

		return await handleApiResponse(response);
	} catch (error) {
		throw new Error(await handleApiError(error));
	}
}
