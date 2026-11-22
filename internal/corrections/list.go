package corrections

const enInstructions = "/en/immigration-refugees-citizenship/corporate/mandate/policies-operational-instructions-agreements/ministerial-instructions/express-entry-rounds/"

const frInstructions = "/fr/immigration-refugies-citoyennete/organisation/mandat/politiques-directives-operationnelles-ententes-accords/instructions-ministerielles/entree-express-rondes/"

// All lists every correction. Add one per mistake, with the value exactly
// as IRCC published it.
var All = []Correction{
	{
		Lang:  French,
		Round: "149",
		Field: "drawNumberURL",
		Wrong: "<a href='" + frInstructions + "invitations-14.html9'>149</a>",
		Right: "<a href='" + frInstructions + "invitations-149.html'>149</a>",
		Why:   "the link's round number was split around .html",
	},
	{
		Lang:  English,
		Round: "34",
		Field: "drawNumberURL",
		Wrong: "<a href='" + enInstructions + "invitations-34.html'>34</a>",
		Right: "",
		Why:   "the English page doesn't exist on canada.ca (404), so the link is dropped",
	},
}
