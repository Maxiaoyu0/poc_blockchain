#!/bin/bash

convertToOrgName () {
	if [ -z $1 ]; then
		echo "No params passed to convertToOrgName"
	else
		if [ $1 -eq 1 ]
    then
      echo "sfeir"
    elif [ $1 -eq 2 ]
    then
      echo "ticketMaster"
    elif [ $1 -eq 3 ]
    then
      echo "bankRoute"
    else
      echo "org3"
    fi
	fi
}

echo "min: $(convertToOrgName 1)"
echo "MAJ: $(convertToOrgName 1 | tr [a-z] [A-Z])"