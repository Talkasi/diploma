REPORT_DIR = report
SLIDES_DIR = slides
MAIN       = report
OUTDIR     = release

SLIDES_SRC = $(firstword $(wildcard $(SLIDES_DIR)/*.pptx))
SLIDES_OUT = $(OUTDIR)/slides.pdf
LIBREOFFICE = libreoffice

all: $(OUTDIR)/$(MAIN).pdf $(SLIDES_OUT)

$(OUTDIR)/$(MAIN).pdf: $(REPORT_DIR)/$(MAIN).pdf
	mkdir -p $(OUTDIR)
	gs -sDEVICE=pdfwrite \
	  -dCompatibilityLevel=1.4 \
	  -dNOPAUSE \
	  -dOptimize=true \
	  -dQUIET \
	  -dBATCH \
	  -dRemoveUnusedFonts=true \
	  -dRemoveUnusedImages=true \
	  -dOptimizeResources=true \
	  -dDetectDuplicateImages \
	  -dCompressFonts=true \
	  -dEmbedAllFonts=true \
	  -dSubsetFonts=true \
	  -dPreserveAnnots=true \
	  -dPreserveMarkedContent=true \
	  -dPreserveOverprintSettings=true \
	  -dPreserveHalftoneInfo=true \
	  -dPreserveOPIComments=true \
	  -dPreserveDeviceN=true \
	  -dMaxInlineImageSize=0 \
	  -sOutputFile=$@ \
	  $(REPORT_DIR)/$(MAIN).pdf

$(REPORT_DIR)/$(MAIN).pdf:
	cd $(REPORT_DIR) && pdflatex -interaction=nonstopmode $(MAIN).tex
	cd $(REPORT_DIR) && biber $(MAIN)
	cd $(REPORT_DIR) && for i in 1 2; do pdflatex -interaction=nonstopmode $(MAIN).tex; done

$(SLIDES_OUT): $(SLIDES_SRC)
	mkdir -p $(OUTDIR)
	$(LIBREOFFICE) --headless --convert-to pdf --outdir $(OUTDIR) $<

.PHONY: clean
clean:
	cd $(REPORT_DIR) && rm -f *.toc *.aux *.out *.gz *.log *.bcf *.blg *.bbl *.xml *.run.xml *.pdf
