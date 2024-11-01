// orca-st.go
//
// go run orca-st.go -f orca.out -options
//
// build the executable:
// go build orca-st.go

package main

import (
    "fmt"                   // general
    "os"                    // open file
    "strings"               // string manipulation
    "strconv"               // string conversion, e.g. int -> string
    "regexp"                // regex
    "flag"                  // argument parser
    )
    
type transitions struct {
    // extraction from SPECSTRINGSTART to SPECSTRINGEND
    State      int          // State number
    Transition string       // for ORCA 6 only
    Energy     string       // energy in cm-1
    Wavelength string       // in nm
    fosc       string       // fosc
}

type states struct {
    // extraction from STATESTRINGSTART to STATESTRINGEND
    State          int      // State number
    Orbs          []string  // [go slice] orbitals, e.g. 54a -> 59a
    Weight        []float64 // [go slice] weight of the individual excitations as float
    OrbsWeight    string    // orbitals and weight in %, e.g. 54a -> 59a (22.1%)
    OrbsWeightLen int       // length of orbitals and weight line
}

// contants for detecting several strings in orca.out
const SPECSTRINGSTART  = "ABSORPTION SPECTRUM VIA TRANSITION ELECTRIC DIPOLE MOMENTS"
const SPECSTRINGEND    = "ABSORPTION SPECTRUM VIA TRANSITION VELOCITY DIPOLE MOMENTS"
const STATESTRINGSTART = "TD-DFT/TDA EXCITED STATES"
const STATESTRINGEND   = "TD-DFT/TDA-EXCITATION SPECTRA"
const NTOSTRING        = "NATURAL TRANSITION ORBITALS"

func openfile(filename string) string{
    // open orca.out and return content as string
    file, err := os.ReadFile(filename)
    // file not found etc.
    if err != nil {
        fmt.Println(err)
    }
    // something is not correct: wrong file, no absorption data, etc.
    if strings.Index(string(file), SPECSTRINGSTART) == -1 || 
       strings.Index(string(file), "Program Version") == -1 {
        fmt.Println("Wrong instruction or file does not contain relevant data. Exit.")
        // exit prg
        os.Exit(0)
    }
    return string(file) 
}

func maketransmap(text string) map[int]transitions{
    // make the transitions map (dict)
    // map[state numer] = struct transitions
    m := make(map[int]transitions)
    
    // manual state count (state number) for ORCA ver. >= 6 
    i := 0
    
    // state number for ORCA ver. < 6 
    stateNumber := 0 
    
    // split text string in lines 
    // maybe needed for windows
    // text = strings.ReplaceAll(text, "\r\n", "\n")
    lines := strings.Split(text, "\n")
    
    // regex detect pattern
    // ORCA ver. >= 6: transition, energy eV, energy cm-1, wavelength, fosc
    // note Submatch(es) in ()    
    transitionPattern := regexp.MustCompile("(\\d+-\\d+[A,B]\\s+->\\s+\\d+-\\d+[A,B])" + 
                                              "\\s+(\\d+.\\d+)\\s+(\\d+.\\d+)\\s+" +
                                              "(\\d+.\\d+)\\s+(\\d+.\\d+)")
    // regex detect pattern                                         
    // ORCA ver. < 6: (space+) state, energy cm-1, wavelength, fosc  
    // note Submatch(es) in ()
    statePattern := regexp.MustCompile("\\s+(\\d+)\\s+(\\d+.\\d+)" + 
                                       "\\s+(\\d+.\\d+)\\s+(\\d+.\\d+)")
    // iterate over single lines
    for _, line := range lines {
        // state (i) + transition + energy (cm-1) + wavelength (nm) + fosc for ORCA ver. >= 6
        // add if regex pattern matches and ORCA >= ver. 6
        if transitionMatch := transitionPattern.FindStringSubmatch(line); 
           transitionMatch != nil {
            i++
            // add to transitions map (dict)
            m[i] = transitions{
                   State:      i, 
                   Transition: transitionMatch[1],
                   Energy:     transitionMatch[3],
                   Wavelength: transitionMatch[4],
                   fosc:       transitionMatch[5],
            }
        }
        // state (from line) + energy (cm-1) + wavelength (nm) + fosc for ORCA ver. >= 6
        //  add if regex pattern matches and ORCA ver. < 6
        if stateMatch := statePattern.FindStringSubmatch(line); stateMatch != nil {
            stateNumber, _ = strconv.Atoi(stateMatch[1])
            // add to transitions map (dict)
            m[stateNumber] = transitions{
                             State:      stateNumber, 
                             Energy:     stateMatch[2],
                             Wavelength: stateMatch[3],
                             fosc:       stateMatch[4],
            }
        }
    }
    // return map (dict)
    return m
}

func makestatesmap(text string, threshold float64, nto bool) map[int]states{
    // make thestates map (dict)
    
    // check if threshold is in the range 0 to 100 %; otherwise set to zero
    if threshold < 0 || threshold > 100 {
        fmt.Println("Warning. Threshold must be between 0 and 100%.")
        fmt.Println("Threshold is set to zero (0).")
        threshold = 0
    } 
    
    // map[state numer] = struct states
    m := make(map[int]states)
    
    // define some variables
    stateNumber := 0            // state number
    trans       := []string{}   // [go slice] orbitals, e.g. 54a -> 59a
    weight      := []float64{}  // [go slice] weight of the individual excitations as float
    orbsweight  := []string{}   // [go slice] orbitals and weight in %, e.g. 54a -> 59a (22.1%)
    var owl    string           // joined items from 'orbsweight' slice (list)
    var owllen int              // length of the joined items in 'owl'
    var statePattern      *regexp.Regexp  // statePattern regex
    var transitionPattern *regexp.Regexp  // transitionPattern regex
    
    if nto == true {
    // check if NTO transitions are requested
        if strings.Index(string(text), NTOSTRING) == -1 {
            // no NTO in orca.out detected -> exit prg
            fmt.Println("File does not contain NTO data. Exit.")
            os.Exit(0)
        } else {
            // orca.out with NTO regex detect patterns
            // note Submatch(es) in ()   
            // state number
            statePattern = regexp.MustCompile(NTOSTRING + "?[^0-9]+STATE\\s+(\\d+)")
            // transition and weight, e.g. 54a -> 59a and 0.45621331; note 'n=' for NTO
            transitionPattern = regexp.MustCompile("(\\d+[a,b]\\s+->\\s+\\d+[a,b])\\s+: n=\\s+([\\d.]+)")
        }
    } else {
        // orca.out (with no NTO) regex detect patterns
        // note Submatch(es) in ()   
        // state number
        statePattern = regexp.MustCompile("STATE\\s+(\\d+)")
        // transition and weight, e.g. 54a -> 59a and 0.45621331
        transitionPattern = regexp.MustCompile("(\\d+[a,b]\\s+->\\s+\\d+[a,b])\\s+:\\s+([\\d.]+)")
    }
    
    // split text string in lines 
    // maybe needed for windows 
    // text = strings.ReplaceAll(text, "\r\n", "\n")
    lines := strings.Split(text, "\n")
    
    // iterate over single lines
    for _, line := range lines {
        if stateMatch := statePattern.FindStringSubmatch(line); stateMatch != nil {
            // get state number from line
            // -> STATE 1 <-
            //    151b -> 157b  :     0.031880
            //    ...
            stateNumber, _ = strconv.Atoi(stateMatch[1])
            // set the following vars to 'nil' or '""' or '0' before processing a new line 
            trans      = nil
            weight     = nil
            orbsweight = nil
            owl        = ""
            owllen     = 0
        } else if stateNumber != 0 {
            // get transitions and weights for the state number
            //    STATE 1
            // -> 151b -> 157b  :     0.031880 <-
            // -> ... <-
            if transitionMatch := transitionPattern.FindStringSubmatch(line); 
               transitionMatch != nil {
                // get weights (string to float) and convert to % 
                w, _ := strconv.ParseFloat(transitionMatch[2], 64)
                w = w * 100
                // check threshold from -t option, only add transitions >= threshold
                if w >= threshold { 
                    // transitions, e.g. 151b -> 157
                    trans = append(trans,transitionMatch[1]) 
                    // weights in % as float
                    weight = append(weight, w)
                    // weights in % converted to string
                    ow := fmt.Sprintf("%s (%.1f%%), ",transitionMatch[1], w)
                    // [go slice] orbitals and weight in %, e.g. 54a -> 59a (22.1%)
                    orbsweight = append(orbsweight, ow)
                    // joined items from 'orbsweight' slice (list)
                    owl = strings.TrimRight(strings.Join(orbsweight,""),", ")
                    // length of the joined items in 'owl'
                    owllen = len(owl)
                }
                 // add to states map (dict)
                m[stateNumber] = states {
                                 State:         stateNumber,
                                 Orbs:          trans,
                                 Weight:        weight,
                                 OrbsWeight:    owl,
                                 OrbsWeightLen: owllen,
                }
            }
        }
    }
    // return map (dict)    
    return m
}

func printtabletrans(map1 map[int]transitions, map2 map[int]states){
    // prints the md table with ORCA ver. >= 6 Transition
    // note the 'Transition' column, otherwise equal to 'func printtable'
    header1 := "| State | Transition      | Energy (cm⁻¹) | Wavelength (nm) | fosc         | Selected transitions"
    header2 := "|-------|-----------------|---------------|-----------------|--------------|---------------------"
    // width of the 'Selected transitions' column is adpated to the 
    // length of the joined items in map.OrbsWeight = map.OrbsWeightLen
    if maxlen(map2) - 20 >= 0 {
        fmt.Println(header1 + strings.Repeat(" ", maxlen(map2) - 20) + " |")
        fmt.Println(header2 + strings.Repeat("-", maxlen(map2) - 19) + "|")
    } else {
        fmt.Println(header1 + "|")
        fmt.Println(header2 + "|")
    }
    // iterate over transitions (map1) and states (map2) maps 
    for i := 1; i <= len(map1); i++ {
        // fill columns with values from maps
        fmt.Printf("|  %4d | %15s | %13s | %15s | %12s | %13s\n", 
                   i, 
                   map1[i].Transition, 
                   map1[i].Energy, 
                   map1[i].Wavelength, 
                   map1[i].fosc,
                   // fill with extra spaces for aligment with 'Selected transitions'
                   map2[i].OrbsWeight + 
                           strings.Repeat(" ", maxlen(map2) - map2[i].OrbsWeightLen) +
                           " |",
                   )
    }
    return
}

func printtable(map1 map[int]transitions, map2 map[int]states){
    // prints the md table without ORCA ver. >= 6 Transition
    // no 'Transition' column, otherwise equal to 'func printtabletrans'
    header1 := "| State | Energy (cm⁻¹) | Wavelength (nm) | fosc         | Selected transitions"
    header2 := "|-------|---------------|-----------------|--------------|---------------------"
    // width of the 'Selected transitions' column is adpated to the 
    // length of the joined items in map.OrbsWeight = map.OrbsWeightLen
    if maxlen(map2) - 20 >= 0 {
        fmt.Println(header1 + strings.Repeat(" ", maxlen(map2) - 20) + " |")
        fmt.Println(header2 + strings.Repeat("-", maxlen(map2) - 19) + "|")
    } else {
        fmt.Println(header1 + "|")
        fmt.Println(header2 + "|")
    }
    // iterate over transitions (map1) and states (map2) maps 
    for i := 1; i <= len(map1); i++ {
        // fill columns with values from maps
        fmt.Printf("|  %4d | %13s | %15s | %12s | %13s\n", 
                   i,
                   map1[i].Energy, 
                   map1[i].Wavelength, 
                   map1[i].fosc,
                   // fill with extra spaces for aligment with 'Selected transitions'
                   map2[i].OrbsWeight + 
                           strings.Repeat(" ", maxlen(map2) - map2[i].OrbsWeightLen) + 
                           " |",
                   )
    }
    return
}

func maxlen(map2 map[int]states) int {
    // helper for 'func printtabletrans' and 'func printtable'
    // get the size of the largest 'transitions (percentages)' value: m.OrbsWeightLen
    // needed for the 'Selected transitions' column for '---' and whitespaces
    var maxLen int
    // iterate over states map (map2)
    for _, state := range map2 {
        if state.OrbsWeightLen > maxLen {
            maxLen = state.OrbsWeightLen
        }
    }
    // minimal length
    if maxLen == 0 {
        maxLen = 19
    }
    
    // return the size of the largest 'transitions (percentages)'
    return maxLen
}

func main() {

    // define flags (arguments)
    // filename
    fileFlag := flag.String("f", "orca.out", "orca.out file")
    // treshold for printing transitions 
    threshFlag := flag.Float64("t", 0, "Transitions below the threshold value will not be printed.")
    // include ORCA ver. >= 6 'Transition'
    transFlag := flag.Bool("tr", false, "Print Transition for orca_6.out files.")
    // include NTO instead of regular states
    ntoFlag := flag.Bool("nto", false, "Print NTO transitions.")
    // parse all flags
    flag.Parse()
    
    // open orca.out (prg -f <orca.out>) and get the file content
    content := openfile(*fileFlag)
    
    // extraction of the releveant part from orca.out for the transitions map (dict)
    start_tr := strings.Index(content, SPECSTRINGSTART)
    end_tr := strings.Index(content, SPECSTRINGEND)
    text_tr := content[start_tr : end_tr]
    // generate the transitions map (dict)
    map_tr := maketransmap(text_tr)
    
    // extraction of the releveant part from orca.out for the states map (dict)
    start_st := strings.Index(content, STATESTRINGSTART)
    end_st := strings.Index(content, STATESTRINGEND)
    text_st := content[start_st : end_st]
    // generate the states map (dict)
    // include threshold and nto flags
    map_st := makestatesmap(text_st, *threshFlag, *ntoFlag)
    
    // generate the md table
    if *transFlag == true {
        // include ORCA ver. >= 6 'Transition'
        printtabletrans(map_tr,map_st)
    } else {
        // do not include ORCA ver. >= 6 'Transition'
        printtable(map_tr,map_st)
    }
}
