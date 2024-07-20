package query

import "testing"

func TestIsValidFilter(t *testing.T) {
	type args struct {
		columns []string
		filter  string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid filter when filter is empty",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "",
			},
			want: true,
		},
		{
			name: "valid filter when columns is empty",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id='6f7c13c8-9c6a-432f-a5f6-80a0a1bd29eb'",
			},
			want: true,
		},
		{
			name: "valid filter",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id>1 AND first_name='Alice'",
			},
			want: true,
		},
		{
			name: "valid filter with one operator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1",
			},
			want: true,
		},
		{
			name: "valid filter with two operators",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id<1 AND first_name='Alice' OR last_name='Smith'",
			},
			want: true,
		},
		{
			name: "valid filter with three operators",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 AND first_name='Alice' AND last_name='Smith' OR email='alice@mail.com'",
			},
			want: true,
		},
		{
			name: "invalid filter with bad pair building",
			args: args{
				columns: []string{"id"},
				filter:  "id",
			},
			want: false,
		},
		{
			name: "invalid filter with extra operator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 AND first_name='Alice' AND",
			},
			want: false,
		},
		{
			name: "invalid filter with bad column name",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 AND first_name='Alice' AND name='Smith'",
			},
			want: false,
		},
		{
			name: "invalid filter with no pairs",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id AND first_name='Alice' AND last_name='Smith'",
			},
			want: false,
		},
		{
			name: "invalid filter with operator at the beginning",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "OR id=1 AND ='Alice' AND last_name='Smith'",
			},
			want: false,
		},
		{
			name: "invalid filter with no columns",
			args: args{
				columns: []string{},
				filter:  "id=1 AND first_name='Alice'",
			},
			want: false,
		},
		{
			name: "invalid filter with invalid operator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 LIKE first_name='Alice' AND last_name='Smith'",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidFilter(tt.args.columns, tt.args.filter); got != tt.want {
				t.Errorf("IsValidFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
