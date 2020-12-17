#!/bin/bash

singularity run \
--bind /home/cwilson/compbio/:/compbio \
/home/cwilson/compbio/docker/fastqc_v0_11_9.sif \
fastqc \
--outdir /compbio/analysis/JimCoffman/jcoffman_006.abOnly_cortisol_2018 \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R1_001.fastq.gz \
/compbio/data/JimCoffman/jcoffman_006.abOnly_cortisol_2018/0AB1_S1_LALL_R2_001.fastq.gz \
