// Handles API responses in both legacy and new message formats
export async function handleApiResponse(response) {
    if (!response.ok) throw response;

    const contentType = response.headers.get("Content-Type");
    if (!contentType || !contentType.includes("application/json")) {
        return response;
    }

    const data = await response.json();

    // Check if response is in the new message format
    if (data && typeof data === 'object' && 'ok' in data && 'message' in data) {
        if (!data.ok) {
            throw new Error(data.message?.error || 'Unknown error');
        }
        return data.message;
    }

    // Legacy format - return as is
    return data;
}

// Handles API errors and returns a user-friendly error message
export function handleApiError(error) {
    if (error instanceof Response) {
        switch (error.status) {
            case 401:
                return "Please login to continue";
            case 403:
                return "You don't have permission to do this";
            case 404:
                return "Content not found";
            case 500:
                return "Server error, please try again later";
            default:
                return `Error: ${error.status}`;
        }
    }

    return error.message || "Unknown error occurred";
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
        throw new Error(handleApiError(error));
    }
}
