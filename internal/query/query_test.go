package query

import (
	"testing"
)

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
			name: "valid filter when name has spaces",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "salary", "age"},
				filter:  "first_name='Alice Julie' OR last_name='Smith' and age=20 or salary=1000",
			},
			want: true,
		},
		{
			name: "valid filter when name has spaces and there is more than one operator and spaces",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "salary", "age"},
				filter:  "first_name='Alice Julie'    OR last_name!='Smith' AND age>=20 or salary<=1000 AND email!='this@mail.com'",
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
			name: "valid filter, ignore trailing AND",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 AND first_name='Alice' AND",
			},
			want: true,
		},
		{
			name: "invalid filter, taking into account trailing AND because the space is not ignored",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				filter:  "id=1 AND first_name='Alice' AND ",
			},
			want: false,
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

func TestIsValidSort(t *testing.T) {
	type args struct {
		columns []string
		sort    string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid sort when sort is empty",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "",
			},
			want: true,
		},
		{
			name: "valid sort with one column",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id DESC",
			},
			want: true,
		},
		{
			name: "valid sort with two columns",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id ASC, first_name DESC",
			},
			want: true,
		},
		{
			name: "invalid sort when columns is empty",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id",
			},
			want: false,
		},
		{
			name: "invalid sort with two columns no operator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id,first_name",
			},
			want: false,
		},
		{
			name: "invalid sort with two columns bad separator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id ASC first_name DESC",
			},
			want: false,
		},
		{
			name: "invalid sort with two columns one missing operator",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id ASC, first_name",
			},
			want: false,
		},
		{
			name: "invalid sort with two columns one bad operator name",
			args: args{
				columns: []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				sort:    "id ASC, first_name DES",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidSort(tt.args.columns, tt.args.sort); got != tt.want {
				t.Errorf("IsValidSort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidFields(t *testing.T) {
	type args struct {
		fields  []string
		partial string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid fields when partial is empty",
			args: args{
				fields:  []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				partial: "",
			},
			want: true,
		},
		{
			name: "valid fields with spaces",
			args: args{
				fields:  []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				partial: "id,   first_name, last_name,email",
			},
			want: true,
		},
		{
			name: "valid fields with no spaces",
			args: args{
				fields:  []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				partial: "id,first_name,last_name",
			},
			want: true,
		},
		{
			name: "invalid fields when fields is empty",
			args: args{
				fields:  []string{},
				partial: "id, first_name, last_name",
			},
			want: false,
		},
		{
			name: "invalid fields when name is missing",
			args: args{
				fields:  []string{"id", "first_name", "last_name", "email", "created_at", "updated_at"},
				partial: "id,first_name,last_name,no_valid",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidFields(tt.args.fields, tt.args.partial); got != tt.want {
				t.Errorf("IsValidFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefixFilterFields(t *testing.T) {
	type args struct {
		filter string
		prefix string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "prefix filter with no filter",
			args: args{
				filter: "",
				prefix: "users.",
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "prefix filter with no prefix",
			args: args{
				filter: "id=1 AND first_name='Alice' or last_name='Smith'",
				prefix: "",
			},
			want:    "id=1 AND first_name='Alice' or last_name='Smith'",
			wantErr: false,
		},
		{
			name: "prefix filter with spaces",
			args: args{
				filter: "id=1 AND first_name='Alice' or last_name='Smith'",
				prefix: "users.",
			},
			want:    "users.id=1 AND users.first_name='Alice' or users.last_name='Smith'",
			wantErr: false,
		},
		{
			name: "prefix filter with spaces and no prefix",
			args: args{
				filter: "id=1 AND first_name='Alice' or last_name='Smith'",
				prefix: " ",
			},
			want:    "id=1 AND first_name='Alice' or last_name='Smith'",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PrefixFilterFields(tt.args.filter, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("PrefixFilterFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PrefixFilterFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
