#!/bin/sh -e
fn="minimal-info-unique-tests.$(date +%Y%m%d)"
wget -O"$fn" \
  https://bioportal.salud.gov.pr/api/administration/reports/minimal-info-unique-tests
ln -sf "$fn" minimal-info-unique-tests
