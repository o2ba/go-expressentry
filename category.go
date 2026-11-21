package expressentry

// Category groups rounds by who they targeted. A round whose name matches
// no category is [Other]; IRCC's wording stays in [Draw.Name].
type Category string

const (
	CEC        Category = "cec"
	French     Category = "french"
	PNP        Category = "pnp"
	Healthcare Category = "healthcare"
	Trades     Category = "trades"
	Transport  Category = "transport"
	Managers   Category = "managers"
	Physicians Category = "physicians"
	Military   Category = "military"
	STEM       Category = "stem"
	Education  Category = "education"
	General    Category = "general"
	Other      Category = "other"
)

var categoryNames = map[Category]Name{
	CEC:        {EN: "Canadian Experience Class", FR: "Catégorie de l’expérience canadienne"},
	French:     {EN: "French", FR: "Français"},
	PNP:        {EN: "Provincial Nominee Program", FR: "Programme des candidats des provinces"},
	Healthcare: {EN: "Healthcare", FR: "Santé et services sociaux"},
	Trades:     {EN: "Trades", FR: "Métiers spécialisés"},
	Transport:  {EN: "Transport", FR: "Transports"},
	Managers:   {EN: "Senior managers", FR: "Cadres supérieurs"},
	Physicians: {EN: "Physicians", FR: "Médecins"},
	Military:   {EN: "Military", FR: "Militaires"},
	STEM:       {EN: "STEM", FR: "STIM"},
	Education:  {EN: "Education", FR: "Éducation"},
	General:    {EN: "General", FR: "Général"},
	Other:      {EN: "Other", FR: "Autres"},
}

// Categories returns every Category.
func Categories() []Category {
	return []Category{CEC, French, PNP, Healthcare, Trades, Transport, Managers, Physicians, Military, STEM, Education, General, Other}
}

// Valid reports whether c is one of [Categories].
func (c Category) Valid() bool {
	_, ok := categoryNames[c]
	return ok
}

// Name returns the category's name in English and French, or an empty Name
// when c isn't valid.
func (c Category) Name() Name {
	return categoryNames[c]
}
