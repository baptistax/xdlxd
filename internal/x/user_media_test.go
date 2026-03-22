package x

import "testing"

func TestShouldFetchProfilePage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		maxPages int
		want     bool
	}{
		{name: "unlimited first page", page: 1, maxPages: 0, want: true},
		{name: "unlimited later page", page: 500, maxPages: -1, want: true},
		{name: "limited page within cap", page: 3, maxPages: 3, want: true},
		{name: "limited page past cap", page: 4, maxPages: 3, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldFetchProfilePage(tt.page, tt.maxPages); got != tt.want {
				t.Fatalf("shouldFetchProfilePage(%d, %d) = %v, want %v", tt.page, tt.maxPages, got, tt.want)
			}
		})
	}
}

func TestExtractUserRestIDPrefersRequestedUserNode(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"tweet": map[string]any{
				"result": map[string]any{
					"rest_id": "tweet-123",
					"legacy":  map[string]any{},
				},
			},
			"user": map[string]any{
				"result": map[string]any{
					"__typename": "User",
					"rest_id":    "user-456",
					"legacy": map[string]any{
						"screen_name": "ExampleUser",
					},
				},
			},
		},
	}

	if got := extractUserRestID(payload, "@exampleuser"); got != "user-456" {
		t.Fatalf("extractUserRestID() = %q, want %q", got, "user-456")
	}
}

func TestExtractUserRestIDHandlesNestedResultWrapper(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"user": map[string]any{
				"result": map[string]any{
					"result": map[string]any{
						"__typename": "User",
						"rest_id":    "user-789",
						"legacy": map[string]any{
							"screen_name": "nested",
						},
					},
				},
			},
		},
	}

	if got := extractUserRestID(payload, "nested"); got != "user-789" {
		t.Fatalf("extractUserRestID() = %q, want %q", got, "user-789")
	}
}

func TestExtractNextCursorPrefersBottomTimelineCursor(t *testing.T) {
	payload := map[string]any{
		"data": map[string]any{
			"user": map[string]any{
				"result": map[string]any{
					"timeline_v2": map[string]any{
						"timeline": map[string]any{
							"instructions": []any{
								map[string]any{
									"type": "TimelineAddEntries",
									"entries": []any{
										map[string]any{
											"entryId": "cursor-top-1",
											"content": map[string]any{
												"entryType":  "TimelineTimelineCursor",
												"cursorType": "Top",
												"value":      "top-cursor",
											},
										},
										map[string]any{
											"entryId": "cursor-bottom-1",
											"content": map[string]any{
												"entryType":  "TimelineTimelineCursor",
												"cursorType": "Bottom",
												"value":      "bottom-cursor",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"metadata": map[string]any{
				"randomCursor": "wrong-cursor",
			},
		},
	}

	if got := extractNextCursor(payload); got != "bottom-cursor" {
		t.Fatalf("extractNextCursor() = %q, want %q", got, "bottom-cursor")
	}
}

func TestExtractNextCursorFallsBackToShowMoreCursor(t *testing.T) {
	payload := map[string]any{
		"instructions": []any{
			map[string]any{
				"type": "TimelineReplaceEntry",
				"entry": map[string]any{
					"entryId": "cursor-showmore-1",
					"content": map[string]any{
						"entryType":  "TimelineTimelineCursor",
						"cursorType": "ShowMore",
						"value":      "show-more-cursor",
					},
				},
			},
		},
		"someCursorLikeKey": "should-not-win",
	}

	if got := extractNextCursor(payload); got != "show-more-cursor" {
		t.Fatalf("extractNextCursor() = %q, want %q", got, "show-more-cursor")
	}
}
