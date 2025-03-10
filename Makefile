typst_compile=typst compile
TYPST_HEADERS=pdfs/header.typ
PROPOSAL_DEPS=pdfs/proposal/proposal.typ $(TYPST_HEADERS)

proposal: $(PROPOSAL_DEPS)
	$(typst_compile) --root pdfs/ $< pdfs/$@.pdf

.PHONY: proposal
