package handlers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

func TestScanRasterRefs(t *testing.T) {
	entities := []scanEntity{
		{
			entity: entityProject, id: "p1", label: "Проект Альфа",
			href: "/projects/web/alpha",
			fields: []scanField{
				{name: "image", value: "/uploads/projects/alpha.png"},
				{name: "translations.en.body", value: "<img src=\"https://api.example.com/uploads/projects/alpha.png\"> и ![b](/uploads/shared/beta.png)"},
			},
		},
		{
			entity: entityPage, id: "pg1", label: "Статья",
			fields: []scanField{
				{name: "image", value: "/uploads/shared/beta.PNG?v=2"},
				// тот же файл в том же поле повторно — дедуплицируется
				{name: "translations.ru.body", value: "/uploads/shared/gamma.jpg и снова /uploads/shared/gamma.JPG"},
				// чужие URL и склейки с именем файла не считаются ссылками
				// на медиатеку; точка в конце предложения — считается
				{
					name: "translations.en.intro",
					value: "<img src=\"https://cdn.example.org/wp-content/uploads/2023/05/wphero.jpg\"> " +
						"/uploads/shared/logo.png.bak /uploads/shared/logo.pngx x/uploads/shared/nested.png " +
						"Смотрите /uploads/shared/end.png.",
				},
			},
		},
		{
			entity: entityService, id: "s1", label: "Сервис",
			fields: []scanField{
				{name: "translations.en.body", value: "<p><img src=\"/uploads/shared/delta.jpeg\"></p>"},
			},
		},
		{
			entity: entityLegal, id: "l1", label: "privacy",
			fields: []scanField{
				{name: "translations.en.sections[0].body", value: "<p>без картинок</p>"},
			},
		},
	}

	got := scanRasterRefs(entities)

	if len(got) != 5 {
		t.Fatalf("want 5 unique files, got %d: %v", len(got), got)
	}

	// Чужие URL (путь до /uploads/ продолжается) и склейки с именем не индексируются.
	for _, phantom := range []string{
		"/uploads/2023/05/wphero.jpg",
		"/uploads/shared/logo.png",
		"/uploads/shared/nested.png",
	} {
		if _, ok := got[phantom]; ok {
			t.Errorf("%q must not be indexed", phantom)
		}
	}

	// Точка-конец предложения после расширения не мешает совпадению.
	if end := got["/uploads/shared/end.png"]; len(end) != 1 || end[0].Entity != entityPage {
		t.Errorf("end.png = %+v, want single page usage", end)
	}

	alpha := got["/uploads/projects/alpha.png"]
	if len(alpha) != 2 {
		t.Fatalf("alpha: want 2 usages, got %d", len(alpha))
	}
	if alpha[0].Entity != entityProject || alpha[0].Field != "image" {
		t.Errorf("alpha usage[0] = %+v", alpha[0])
	}
	if alpha[1].Entity != entityProject || alpha[1].Field != "translations.en.body" {
		t.Errorf("alpha usage[1] = %+v", alpha[1])
	}
	if alpha[0].Href != "/projects/web/alpha" {
		t.Errorf("alpha usage[0].href = %q", alpha[0].Href)
	}

	beta := got["/uploads/shared/beta.png"]
	// сортировка по entity: страница (image), затем проект (en.body);
	// .PNG с query-суффиксом нормализуется к тому же файлу
	if len(beta) != 2 {
		t.Fatalf("beta: want 2 usages, got %d", len(beta))
	}
	if beta[0].Entity != entityPage || beta[0].Field != "image" {
		t.Errorf("beta usage[0] = %+v", beta[0])
	}
	if beta[1].Entity != entityProject || beta[1].Field != "translations.en.body" {
		t.Errorf("beta usage[1] = %+v", beta[1])
	}

	// расширение нормализуется к нижнему регистру: .jpg и .JPG — один файл
	gamma := got["/uploads/shared/gamma.jpg"]
	if len(gamma) != 1 || gamma[0].Entity != entityPage {
		t.Errorf("gamma = %+v, want single page usage", gamma)
	}

	delta := got["/uploads/shared/delta.jpeg"]
	if len(delta) != 1 || delta[0].Entity != entityService {
		t.Errorf("delta = %+v, want single service usage", delta)
	}
}

func TestReplaceRasterRefInValue(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   string
		target string
	}{
		{
			"relative url",
			"<img src=\"/uploads/a/one.png\">",
			"<img src=\"/uploads/a/one.webp\">",
			"/uploads/a/one.png",
		},
		{
			"jpeg matches, other jpg untouched",
			"<img src=\"/uploads/a/two.jpg\"> ![x](/uploads/a/three.JPEG)",
			"<img src=\"/uploads/a/two.jpg\"> ![x](/uploads/a/three.webp)",
			"/uploads/a/three.jpeg",
		},
		{
			"absolute url keeps host and query",
			"https://api.example.com/uploads/a/one.png?v=1#frag",
			"https://api.example.com/uploads/a/one.webp?v=1#frag",
			"/uploads/a/one.png",
		},
		{
			"uppercase extension matches",
			"/uploads/a/ONE.PNG",
			"/uploads/a/ONE.webp",
			"/uploads/a/ONE.png",
		},
		{
			"other file untouched",
			"/uploads/a/two.png /uploads/a/one.webp",
			"/uploads/a/two.png /uploads/a/one.webp",
			"/uploads/a/one.png",
		},
		{
			"path prefix must not match",
			"/uploads/a/one-two.png",
			"/uploads/a/one-two.png",
			"/uploads/a/one.png",
		},
		{
			"foreign url with /uploads/ segment untouched",
			"<img src=\"https://cdn.example.org/wp-content/uploads/2023/05/hero.jpg\">",
			"<img src=\"https://cdn.example.org/wp-content/uploads/2023/05/hero.jpg\">",
			"/uploads/2023/05/hero.jpg",
		},
		{
			"filename continuation .bak untouched",
			"/uploads/a/logo.png.bak",
			"/uploads/a/logo.png.bak",
			"/uploads/a/logo.png",
		},
		{
			"filename continuation .pngx untouched",
			"/uploads/a/logo.pngx",
			"/uploads/a/logo.pngx",
			"/uploads/a/logo.png",
		},
		{
			"glued into longer relative path untouched",
			"x/uploads/a/one.png",
			"x/uploads/a/one.png",
			"/uploads/a/one.png",
		},
		{
			"sentence dot after extension still replaced",
			"see /uploads/a/one.png.",
			"see /uploads/a/one.webp.",
			"/uploads/a/one.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceRasterRefInValue(tt.in, tt.target); got != tt.want {
				t.Errorf("replaceRasterRefInValue(%q, %q) = %q, want %q", tt.in, tt.target, got, tt.want)
			}
		})
	}
}

func TestRasterWebPPath(t *testing.T) {
	for in, want := range map[string]string{
		"/uploads/x/shot.png":  "/uploads/x/shot.webp",
		"/uploads/x/SHOT.PNG":  "/uploads/x/SHOT.webp",
		"/uploads/x/shot.jpg":  "/uploads/x/shot.webp",
		"/uploads/x/shot.JPEG": "/uploads/x/shot.webp",
		"/uploads/x/logo.svg":  "/uploads/x/logo.svg",
		"/uploads/x/anim.gif":  "/uploads/x/anim.gif",
		"/uploads/x/shot.webp": "/uploads/x/shot.webp",
	} {
		if got := rasterWebPPath(in); got != want {
			t.Errorf("rasterWebPPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeRasterRefPath(t *testing.T) {
	// Канонизация: нижний регистр расширения, Clean сегментов.
	canon := map[string]string{
		"/uploads/a/b.png":         "/uploads/a/b.png",
		"/uploads/photo.PNG":       "/uploads/photo.png",
		"/uploads/my-file(1).png":  "/uploads/my-file(1).png",
		"/uploads/x/c.jpg":         "/uploads/x/c.jpg",
		"/uploads/x/c.jpeg":        "/uploads/x/c.jpeg",
		"/uploads/a/b/../../c.png": "/uploads/c.png",
		"/uploads/a//d.PNG":        "/uploads/a/d.png",
		"/uploads/ONE.JPEG":        "/uploads/ONE.jpeg",
	}
	for in, want := range canon {
		gotPath, err := normalizeRasterRefPath(in)
		if err != nil {
			t.Errorf("normalizeRasterRefPath(%q) unexpected error: %v", in, err)
			continue
		}
		if gotPath != want {
			t.Errorf("normalizeRasterRefPath(%q) = %q, want %q", in, gotPath, want)
		}
	}

	invalid := []string{
		"",
		"uploads/a.png",
		"https://example.com/uploads/a.png",
		"/uploads/a.gif",
		"/uploads/a.svg",
		"/uploads/a.webp",
		"/uploads/.hidden.png",
		"/uploads/.png",
	}
	for _, p := range invalid {
		if _, err := normalizeRasterRefPath(p); err == nil {
			t.Errorf("normalizeRasterRefPath(%q) must fail", p)
		}
	}
}

func TestLegalTranslationFields(t *testing.T) {
	tl := models.LegalTranslations{
		"en": {
			Label: "Privacy <img src=\"/uploads/legal/label.png\">",
			Title: "Policy",
			Sections: []models.LegalSection{
				{Title: "Intro <img src=\"/uploads/legal/t.png\">", Body: "<p>/uploads/legal/b.png</p>"},
			},
		},
	}

	fields := legalTranslationFields(tl)
	names := make(map[string]string, len(fields))
	for _, f := range fields {
		names[f.name] = f.value
	}

	want := map[string]string{
		"translations.en.label":             "Privacy <img src=\"/uploads/legal/label.png\">",
		"translations.en.title":             "Policy",
		"translations.en.sections[0].title": "Intro <img src=\"/uploads/legal/t.png\">",
		"translations.en.sections[0].body":  "<p>/uploads/legal/b.png</p>",
	}
	if len(names) != len(want) {
		t.Fatalf("fields = %v", names)
	}
	for name, value := range want {
		if names[name] != value {
			t.Errorf("field %q = %q, want %q", name, names[name], value)
		}
	}

	// Замена через set доходит до модели.
	for _, f := range fields {
		f.set(strings.ReplaceAll(f.value, ".png", ".webp"))
	}
	if tl["en"].Label != "Privacy <img src=\"/uploads/legal/label.webp\">" {
		t.Errorf("label = %q", tl["en"].Label)
	}
	if tl["en"].Sections[0].Title != "Intro <img src=\"/uploads/legal/t.webp\">" {
		t.Errorf("section title = %q", tl["en"].Sections[0].Title)
	}
	if tl["en"].Sections[0].Body != "<p>/uploads/legal/b.webp</p>" {
		t.Errorf("section body = %q", tl["en"].Sections[0].Body)
	}
}

func TestApplyRasterReplacementsPartialFailure(t *testing.T) {
	ok := &models.Page{ID: "ok", Slug: "ok", Image: "/uploads/a/one.png"}
	broken := &models.Page{ID: "broken", Slug: "broken", Image: "/uploads/a/two.png"}
	untouched := &models.Page{ID: "clean", Slug: "clean", Image: "/uploads/a/three.png"}

	entities := []scanEntity{
		{
			entity: entityPage, id: ok.ID,
			fields: []scanField{{name: "image", value: ok.Image, set: func(v string) { ok.Image = v }}},
			save:   func(context.Context) error { return nil },
		},
		{
			entity: entityPage, id: broken.ID,
			fields: []scanField{{name: "image", value: broken.Image, set: func(v string) { broken.Image = v }}},
			save:   func(context.Context) error { return errors.New("db down") },
		},
		{
			entity: entityPage, id: untouched.ID,
			fields: []scanField{{name: "image", value: untouched.Image, set: func(v string) { untouched.Image = v }}},
			save: func(context.Context) error {
				t.Error("entity without targets must not be saved")
				return nil
			},
		},
	}

	updated, err := applyRasterReplacements(
		context.Background(), entities,
		map[string]bool{"/uploads/a/one.png": true, "/uploads/a/two.png": true},
	)
	if err == nil {
		t.Fatal("want first save error to surface")
	}
	if updated != 1 {
		t.Errorf("updated = %d, want 1 (only the entity saved before the failure)", updated)
	}
	if ok.Image != "/uploads/a/one.webp" || broken.Image != "/uploads/a/two.webp" {
		t.Errorf("fields must be rewritten even when the save fails: %q %q", ok.Image, broken.Image)
	}
	if untouched.Image != "/uploads/a/three.png" {
		t.Errorf("untouched entity changed: %q", untouched.Image)
	}
}

func TestCollectScannableReplaceFlow(t *testing.T) {
	project := &models.Project{
		ID: "p1", Slug: "alpha", Image: "/uploads/projects/alpha.png",
		Translations: models.Translations{"ru": {"body": "<p>/uploads/projects/alpha.png</p>"}},
	}
	page := &models.Page{ID: "pg1", Slug: "post", Image: "/uploads/post.jpg"}

	var saved []string
	entities := []scanEntity{
		{
			entity: entityProject, id: project.ID,
			fields: []scanField{
				{name: "image", value: project.Image, set: func(v string) { project.Image = v }},
				{name: "translations.ru.body", value: project.Translations["ru"]["body"], set: func(v string) { project.Translations["ru"]["body"] = v }},
			},
			save: func(context.Context) error { saved = append(saved, "project"); return nil },
		},
		{
			entity: entityPage, id: page.ID,
			fields: []scanField{{name: "image", value: page.Image, set: func(v string) { page.Image = v }}},
			save:   func(context.Context) error { saved = append(saved, "page"); return errors.New("must not be saved") },
		},
	}

	// Сканируем, заменяем найденный путь, сохраняем только изменённое.
	usages := scanRasterRefs(entities)
	target := "/uploads/projects/alpha.png"
	if _, ok := usages[target]; !ok {
		t.Fatalf("target not found in scan: %v", usages)
	}

	updated, err := applyRasterReplacements(context.Background(), entities, map[string]bool{target: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 || !reflect.DeepEqual(saved, []string{"project"}) {
		t.Fatalf("updated=%d saved=%v, want 1 [project]", updated, saved)
	}
	if project.Image != "/uploads/projects/alpha.webp" {
		t.Errorf("image = %q", project.Image)
	}
	if project.Translations["ru"]["body"] != "<p>/uploads/projects/alpha.webp</p>" {
		t.Errorf("body = %q", project.Translations["ru"]["body"])
	}
	if page.Image != "/uploads/post.jpg" {
		t.Errorf("untouched entity changed: %q", page.Image)
	}
}

func TestTranslationsTitle(t *testing.T) {
	if got := translationsTitle(models.Translations{"ru": {"title": "Русский"}, "en": {"title": "English"}}, "slug"); got != "English" {
		t.Errorf("en preferred, got %q", got)
	}
	if got := translationsTitle(models.Translations{"de": {"title": "Deutsch"}, "ru": {"title": "Русский"}}, "slug"); got != "Deutsch" {
		t.Errorf("alphabetical fallback, got %q", got)
	}
	if got := translationsTitle(models.Translations{"en": {"title": "  "}}, "slug"); got != "slug" {
		t.Errorf("fallback on empty title, got %q", got)
	}
}
