/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
use std::{error::Error, io, process};
use std::collections::HashSet;
use std::fs::File;
use std::io::Write;

use csv;
use regex::Regex;

#[derive(Default, Debug, Ord, PartialOrd, Eq, PartialEq, Hash, Clone)]
struct IEEntry {
    iei: i32,
    type_name: String,
    type_ref: String,
    mandatory: bool,
    length_size: i32,
    min_length: i32,
    max_length: i32,
}

const LENGTH_7_OR_11_OR_15: i32 = -1;

#[derive(Default, Debug)]
struct MsgEntry {
    struct_name: String,
    section: String,
    is_gmm: bool,
    msg_type: u8,
    ies: Vec<IEEntry>,
}
pub fn generate_nas() -> Result<(), Box<dyn Error>> {
    let msg_entries = parse_message()?;
    let mut total = String::new();
    let mut all_types = HashSet::new();

    for msg in msg_entries.iter().filter(|msg|!msg.is_gmm) {
        let mut strct = String::new();
        strct.push_str(&format!("struct {} {{\n", msg.struct_name));

        for ie in &msg.ies {
            if ie.mandatory {
                strct.push_str(&format!("\t\t{}: {},\n", &ie.type_name, ie.type_name));
            } else {
                strct.push_str(&format!("\t\t{}: Option<{}>,\n", &ie.type_name, ie.type_name));
            }
            all_types.insert(ie);
        }

        strct.push_str("}\n\n");
        total.push_str(&strct);
    }

    let mut file = File::create("nas_generated.rs")?;
    file.write_all(total.as_bytes())?;

    println!("{:#?}", all_types);
    let regex = Regex::new(&format!("(Table\u{00A0}{})", "9.11.4.10")).unwrap();

    let mut rdr = csv::ReaderBuilder::new()
        .flexible(true)
        .from_reader(File::open("nas/spec.csv")?);

    let mut prev_rdr = None;
    for result in rdr.records() {
        let record = result?;

        if let Some(m) = regex.captures(&record[0]) {
            panic!("{:?}", prev_rdr);
        }
        prev_rdr = Some(record);
    }

    Ok(())
}

fn parse_message() -> Result<Vec<MsgEntry>, Box<dyn Error>> {
    let reg_message_content = Regex::new(r"Table\pZ+(8\..+): (.*) message content").unwrap();
    let reg_message_type = Regex::new(r"Table\pZ+9\.7\..+: Message types for (.*)").unwrap();

    let mut rdr = csv::ReaderBuilder::new()
        .flexible(true)
        .from_reader(File::open("nas/spec.csv")?);

    let mut prev_rdr = None;
    let mut dereg_flag = false;
    let mut iter = rdr.records();

    let mut msg_entries = vec![];
    while let Some(result) = iter.next() {
        let record = result?;
        if record.len() == 1 {
            if let Some(m) = reg_message_content.captures(&record[0]) {
                let section_number = &m[1];
                let message_name = &m[2];
                let struct_name = convert_message_name(message_name, &mut dereg_flag);

                let top_fields = iter.next().unwrap()?;
                if &top_fields[0] != "IEI" ||
                    &top_fields[1] != "Information Element" ||
                    &top_fields[2] != "Type/Reference" ||
                    &top_fields[3] != "Presence" ||
                    &top_fields[4] != "Format" ||
                    &top_fields[5] != "Length" {
                    panic!("Invalid fields");
                }
                let mut ie_entries: Vec<IEEntry> = vec![];
                let mut prev_half = false;
                'skip_ie: while let Some(result) = iter.next() {
                    let ie_fields = result?;
                    // Table done
                    if ie_fields.len() == 1 {
                        break;
                        prev_rdr = Some(ie_fields);
                    }

                    let mut ie = IEEntry::default();
                    let iei = &ie_fields[0];
                    let ie_name = &ie_fields[1];
                    let type_ref = &ie_fields[2];
                    let presence = &ie_fields[3];
                    let format = &ie_fields[4];
                    let length = &ie_fields[5];

                    match presence {
                        "M" => {
                            // mandatory IE
                            if !iei.is_empty() {
                                panic!("IEI must be empty");
                            }
                            ie.mandatory = true;
                            // parse format value
                            match format {
                                "V" | "LV" => {
                                    ie.length_size = 1;
                                }
                                "LV-E" => {
                                    ie.length_size = 2;
                                }
                                _ => panic!("Invalid format {}", &format),
                            }
                        }
                        "O" | "C" => {
                            // not mandatory IE
                            if iei.is_empty() {
                                panic!("IEI must not be empty");
                            }
                            // parse IEI value
                            if iei.len() > 1 && iei.chars().nth(1) == Some('-') {
                                if let Ok(i) = u8::from_str_radix(&iei[..1], 16) {
                                    ie.iei = i as i32;
                                } else {
                                    panic!("Failed to parse IEI value");
                                }
                            } else {
                                if let Ok(i) = u8::from_str_radix(&iei, 16) {
                                    ie.iei = i as i32;
                                } else {
                                    panic!("Failed to parse IEI value");
                                }
                            }
                            // parse format value
                            match format {
                                "TV" | "TLV" => {
                                    ie.length_size = 1;
                                }
                                "TLV-E" => {
                                    ie.length_size = 2;
                                }
                                _ => panic!("Invalid format {}", &format),
                            }
                        }
                        _ => panic!("Invalid presence {}", &presence),
                    }

                    // parse length field
                    let mut half = false;
                    let len_split: Vec<&str> = length.split("-").collect();

                    if len_split.len() == 1 {
                        // Fixed length
                        if len_split[0] == "1/2" {
                            // half octet IE
                            half = true;
                            ie.length_size = 0;
                        } else if len_split[0] == "7, 11 or 15" {
                            // Special case (PDU address IE)
                            ie.min_length = LENGTH_7_OR_11_OR_15;
                            ie.max_length = LENGTH_7_OR_11_OR_15;
                        } else {
                            if let Ok(i) = len_split[0].parse::<i32>() {
                                ie.min_length = i;
                                ie.max_length = i;
                            } else {
                                panic!("Failed to parse length value");
                            }
                        }
                    } else {
                        // Length range
                        if let Ok(i) = len_split[0].parse::<i32>() {
                            ie.min_length = i;
                        } else {
                            panic!("Failed to parse length value");
                        }

                        if len_split[1] == "n" {
                            // length is not limited
                            ie.max_length = std::i32::MAX;
                        } else {
                            if let Ok(i) = len_split[1].parse::<i32>() {
                                ie.max_length = i;
                            } else {
                                panic!("Failed to parse length value");
                            }
                        }
                    }

                    // Convert IE name text to go type for IE
                    let mut ie_cell = ie_name.trim();
                    let mut words: Vec<&str> = ie_cell.split(' ').collect();
                    let mut type_name = String::new();

                    // Struct names can't begin with a number, so we rename types like 5GSMCause to Cause5GSM
                    if words[0].starts_with('5') {
                        let first_word = words.remove(0);
                        words.push(first_word);
                    }

                    if ie_cell.starts_with("PDU SESSION ") && ie_cell.ends_with(" message identity") {
                        type_name = ie_cell.replace(" ", "").replace("messageidentity", "MessageIdentity");
                    } else {
                        match ie_cell {
                            "Payload container type" => {
                                if ie.mandatory {
                                    type_name = "PayloadContainerType".to_string();
                                } else {
                                    continue 'skip_ie;
                                }
                            }
                            "5GMM STATUS message identity" => {
                                type_name = "STATUSMessageIdentity5GMM".to_string();
                            }
                            "5GSM STATUS message identity" => {
                                type_name = "STATUSMessageIdentity5GSM".to_string();
                            }
                            "5G-GUTI" => {
                                type_name = "GUTI5G".to_string();
                            }
                            "5G-S-TMSI" => {
                                type_name = "TMSI5GS".to_string();
                            }
                            "Authentication parameter RAND (5G authentication challenge)" => {
                                type_name = "AuthenticationParameterRAND".to_string();
                            }
                            "Authentication parameter AUTN (5G authentication challenge)" => {
                                type_name = "AuthenticationParameterAUTN".to_string();
                            }
                            "PDU session ID" => {
                                if ie.iei == 0x12 {
                                    type_name = "PduSessionID2Value".to_string();
                                } else {
                                    type_name = "PDUSessionID".to_string();
                                }
                            }
                            _ => {
                                for word in words {
                                    match word {
                                        "NAS" | "ABBA" | "EAP" | "TAI" | "NSSAI" | "LADN" | "MICO" | "DL" | "UL" | "SMS"
                                        | "DNN" | "TRANSPORT" | "ID" | "5G" | "5GS" | "5GSM" | "5GMM" | "PDU" | "PTI" | "SSC"
                                        | "AMBR" | "RQ" | "EPS" | "SM" | "DN" | "SOR" | "DRX" | "UE" | "GUTI" | "IMEISV" => {
                                            type_name.push_str(word);
                                        }
                                        "S-NSSAI" => {
                                            type_name.push_str("SNSSAI");
                                        }
                                        "Non-3GPP" => {
                                            type_name.push_str("Non3Gpp");
                                        }
                                        _ => {
                                            type_name.push_str(&title_case(&word.replace("'", "").replace("-", "")));
                                        }
                                    }
                                }
                            }
                        }
                    }

                    ie.type_name = type_name;
                    ie.type_ref = type_ref.to_string();

                    if half && prev_half {
                        // Merge IEs have half octet size
                        let prev_ie = ie_entries.last_mut().unwrap();
                        if prev_ie.min_length != 0 {
                            panic!("Merge non half IEs");
                        }
                        if !prev_ie.mandatory || !ie.mandatory {
                            panic!("Merge non mandatory IEs");
                        }
                        if prev_ie.length_size != 0 || ie.length_size != 0 {
                            panic!("Merge IEs has length");
                        }
                        prev_ie.type_name = format!("{}And{}", ie.type_name, prev_ie.type_name);
                        prev_ie.min_length = 1;
                        prev_ie.max_length = 1;
                        prev_ie.length_size = 1;
                        prev_half = false;
                    } else {
                        ie_entries.push(ie);
                        prev_half = half;
                    }
                }
                msg_entries.push(MsgEntry {
                    struct_name,
                    section: section_number.to_string(),
                    ies: ie_entries,
                    ..Default::default()
                });
            }
        } else {
            // TODO !!
        }

        prev_rdr = Some(record);
    }

    Ok(msg_entries)
}

fn convert_message_name(msg_name_in_doc: &str, dereg_flag: &mut bool) -> String {
    let words: Vec<&str> = msg_name_in_doc.split(" ").collect();
    let mut msg_name = String::new();

    match msg_name_in_doc {
        "5GMM STATUS" | "5GMM status" => msg_name = "Status5GMM".to_string(),
        "5GSM STATUS" | "5GSM status" => msg_name = "Status5GSM".to_string(),
        "DEREGISTRATION REQUEST" => {
            if *dereg_flag {
                msg_name = "DeregistrationRequestUETerminatedDeregistration".to_string();
            } else {
                msg_name = "DeregistrationRequestUEOriginatingDeregistration".to_string();
            }
        }
        "Deregistration request (UE terminated)" => {
            msg_name = "DeregistrationRequestUETerminatedDeregistration".to_string();
        }
        "Deregistration request (UE originating)" => {
            msg_name = "DeregistrationRequestUEOriginatingDeregistration".to_string();
        }
        "DEREGISTRATION ACCEPT" => {
            if *dereg_flag {
                msg_name = "DeregistrationAcceptUETerminatedDeregistration".to_string();
            } else {
                msg_name = "DeregistrationAcceptUEOriginatingDeregistration".to_string();
                *dereg_flag = true;
            }
        }
        "Deregistration accept (UE terminated)" => {
            msg_name = "DeregistrationAcceptUETerminatedDeregistration".to_string();
        }
        "Deregistration accept (UE originating)" => {
            msg_name = "DeregistrationAcceptUEOriginatingDeregistration".to_string();
        }
        _ => {
            for word in words {
                match word {
                    "PDU" | "5GMM" | "5GSM" | "5GS" | "UL" | "DL" | "NAS" => msg_name += word,
                    _ => msg_name += &*title_case(word),
                }
            }
        }
    }

    msg_name
}

fn title_case(s: &str) -> String {
    if s.is_empty() {
        return String::new();
    }
    let first_letter = s[..1].to_uppercase();
    let remaining_letters = s[1..].to_lowercase();
    format!("{}{}", first_letter, remaining_letters)
}
