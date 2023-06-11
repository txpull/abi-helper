#!/bin/bash
export $( grep -vE "^(#.*|\s*)$" .test.env )
$@
