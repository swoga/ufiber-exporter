package model

import "encoding/json"

// StringOrBool absorbe un champ JSON que l'API UFiber expose tantot comme
// chaine, tantot comme booleen selon la plateforme.
//
// Les OLT GPON renvoient une chaine pour SfpModule.TxFault, les OLT XGS
// renvoient un booleen. Go etant strictement type, une seule incoherence
// fait echouer le parsing de TOUTE la reponse : sans ce type, aucune
// metrique n'est collectee sur les XGS.
//
// La valeur booleenne est normalisee en "true"/"false" pour rester
// compatible avec le code consommant deja ce champ comme une chaine.
type StringOrBool string

func (s *StringOrBool) UnmarshalJSON(data []byte) error {
	// Une valeur JSON non entouree de guillemets n'est pas une chaine :
	// on tente le booleen avant de se rabattre sur le cas nominal.
	if len(data) > 0 && data[0] != '"' {
		var b bool
		if err := json.Unmarshal(data, &b); err == nil {
			if b {
				*s = "true"
			} else {
				*s = "false"
			}
			return nil
		}
	}

	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = StringOrBool(str)
	return nil
}
