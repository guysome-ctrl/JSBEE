package shared

import (
	"strings"
	"testing"
)

func TestValidatePaperMetadata(t *testing.T) {
	cases := []struct {
		name     string
		location string
		category string
		wantErr  string
		wantLoc  string
		wantCat  string
	}{
		{
			name:     "valid",
			location: "Mumbai, India",
			category: "Economics",
			wantLoc:  "Mumbai, India",
			wantCat:  "Economics",
		},
		{
			name:     "trims whitespace",
			location: "  Berlin  ",
			category: "\tPhysics\n",
			wantLoc:  "Berlin",
			wantCat:  "Physics",
		},
		{
			name:     "empty location",
			location: "",
			category: "Economics",
			wantErr:  "Location is required",
		},
		{
			name:     "blank location",
			location: "   ",
			category: "Economics",
			wantErr:  "Location is required",
		},
		{
			name:     "empty category",
			location: "Mumbai, India",
			category: "",
			wantErr:  "Category is required",
		},
		{
			name:     "blank category",
			location: "Mumbai, India",
			category: "  \t ",
			wantErr:  "Category is required",
		},
		{
			name:     "location too long",
			location: strings.Repeat("a", 121),
			category: "Economics",
			wantErr:  "Location must be at most 120 characters",
		},
		{
			name:     "category too long",
			location: "Mumbai, India",
			category: strings.Repeat("a", 121),
			wantErr:  "Category must be at most 120 characters",
		},
		{
			name:     "location at limit",
			location: strings.Repeat("a", 120),
			category: "Economics",
			wantLoc:  strings.Repeat("a", 120),
			wantCat:  "Economics",
		},
		// 120 emoji are 480 bytes but only 120
		// characters: length is counted in runes.
		{
			name:     "unicode counted in runes",
			location: strings.Repeat("🌍", 120),
			category: "Economics",
			wantLoc:  strings.Repeat("🌍", 120),
			wantCat:  "Economics",
		},
		{
			name:     "unicode over limit",
			location: strings.Repeat("🌍", 121),
			category: "Economics",
			wantErr:  "Location must be at most 120 characters",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loc, cat, err := ValidatePaperMetadata(
				tc.location, tc.category,
			)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf(
						"Expected error %q, got %v",
						tc.wantErr, err,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if loc != tc.wantLoc || cat != tc.wantCat {
				t.Fatalf(
					"Got %q, %q; want %q, %q",
					loc, cat, tc.wantLoc, tc.wantCat,
				)
			}
		})
	}
}

func TestValidateCoreMemberFields(t *testing.T) {
	cases := []struct {
		name     string
		inName   string
		inRole   string
		inDesc   string
		wantErr  string
		wantName string
		wantRole string
		wantDesc string
	}{
		{
			name:     "valid",
			inName:   "A",
			inRole:   "R",
			inDesc:   "D",
			wantName: "A",
			wantRole: "R",
			wantDesc: "D",
		},
		{
			name:     "trims",
			inName:   "  A  ",
			inRole:   "\tR\n",
			inDesc:   "  D  ",
			wantName: "A",
			wantRole: "R",
			wantDesc: "D",
		},
		{
			name:    "empty name",
			inName:  "",
			inRole:  "R",
			inDesc:  "D",
			wantErr: "Name is required",
		},
		{
			name:    "blank role",
			inName:  "A",
			inRole:  "   ",
			inDesc:  "D",
			wantErr: "Role is required",
		},
		{
			name:    "empty desc",
			inName:  "A",
			inRole:  "R",
			inDesc:  "",
			wantErr: "Description is required",
		},
		{
			name:    "name too long",
			inName:  strings.Repeat("a", 121),
			inRole:  "R",
			inDesc:  "D",
			wantErr: "Name must be at most 120 characters",
		},
		{
			name:    "role too long",
			inName:  "A",
			inRole:  strings.Repeat("a", 121),
			inDesc:  "D",
			wantErr: "Role must be at most 120 characters",
		},
		{
			name:     "unicode runes",
			inName:   strings.Repeat("🌍", 120),
			inRole:   "R",
			inDesc:   "D",
			wantName: strings.Repeat("🌍", 120),
			wantRole: "R",
			wantDesc: "D",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, role, desc, err := ValidateCoreMemberFields(
				tc.inName, tc.inRole, tc.inDesc,
			)
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf(
						"Expected error %q, got %v",
						tc.wantErr, err,
					)
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if name != tc.wantName ||
				role != tc.wantRole ||
				desc != tc.wantDesc {
				t.Fatalf(
					"Got %q, %q, %q",
					name, role, desc,
				)
			}
		})
	}
}
