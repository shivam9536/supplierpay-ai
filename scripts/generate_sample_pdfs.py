#!/usr/bin/env python3
"""
Generate sample Purchase Orders and Invoices in varied real-world styles.
Mimics different vendors, industries, and document formats (corporate, letterhead,
minimal, Indian GST format, compact, etc.).
"""
import os
import random
from reportlab.lib import colors
from reportlab.lib.pagesizes import A4
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.units import inch
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, HRFlowable

OUT_DIR = os.path.join(os.path.dirname(__file__), "..", "samples", "pdf")
os.makedirs(OUT_DIR, exist_ok=True)


def fmt_currency(amount, currency="INR"):
    if currency == "INR":
        return random.choice([f"₹{amount:,.2f}", f"Rs. {amount:,.2f}", f"INR {amount:,.2f}"])
    return f"{currency} {amount:,.2f}"


# ─── Realistic document sets (different industries, vendors, formats) ───
DOC_SETS = [
    # Set 1: Corporate IT / Cloud (clean, modern)
    {
        "id": "01_corporate",
        "buyer": {"name": "Ventura Technologies Pvt Ltd", "address": "Tower A, Cyber City\nGurgaon, Haryana 122002", "email": "purchase@venturatech.in", "phone": "+91-124-4567890"},
        "vendor": {"name": "Acme Cloud Solutions", "address": "456 Tech Hub, Andheri East\nMumbai 400069", "email": "billing@acmecloud.com", "gstin": "27AABCU9603R1ZM"},
        "po": {"number": "PO-2026-100", "date": "01/03/2026", "approved": "Rahul Sharma", "terms": "Net 30", "currency": "INR",
               "items": [{"desc": "Cloud Hosting - Monthly", "qty": 1, "rate": 41000, "total": 41000}, {"desc": "Support Services", "qty": 1, "rate": 9000, "total": 9000}], "total": 50000},
        "inv": {"number": "INV-ACME-2026-050", "date": "10/03/2026", "due": "09/04/2026", "po_ref": "PO-2026-100", "currency": "INR",
                "items": [{"desc": "Cloud Hosting - March", "qty": 1, "rate": 41000, "total": 41000}, {"desc": "Support Services", "qty": 1, "rate": 9000, "total": 9000}],
                "subtotal": 50000, "tax_pct": 18, "tax": 9000, "total": 59000, "bank": "HDFC Bank A/c 1234567890 IFSC HDFC0001234"},
    },
    # Set 2: Trading / wholesale (dense, compact, many line items)
    {
        "id": "02_trading",
        "buyer": {"name": "Sagar Trading Co.", "address": "Plot 12, Kandivali Industrial Area\nMumbai 400067", "email": "accounts@sagartrading.com", "phone": "022-28901234"},
        "vendor": {"name": "Northern Electronics Ltd", "address": "Sector 18, Udyog Vihar\nGurgaon 122015", "email": "sales@northernelectronics.in", "gstin": "06AABCN1234A1Z5"},
        "po": {"number": "PO/MUM/2026/0892", "date": "05-Mar-2026", "approved": "M. Patel", "terms": "15 days", "currency": "INR",
               "items": [{"desc": "HDMI Cables 2m", "qty": 500, "rate": 85, "total": 42500}, {"desc": "USB Hub 4-Port", "qty": 200, "rate": 320, "total": 64000}, {"desc": "Keyboard (Wired)", "qty": 150, "rate": 450, "total": 67500}], "total": 174000},
        "inv": {"number": "BL-2026-456", "date": "12/03/2026", "due": "27/03/2026", "po_ref": "PO/MUM/2026/0892", "currency": "INR",
                "items": [{"desc": "HDMI Cables 2m", "qty": 500, "rate": 85, "total": 42500}, {"desc": "USB Hub 4-Port", "qty": 200, "rate": 320, "total": 64000}, {"desc": "Keyboard (Wired)", "qty": 150, "rate": 450, "total": 67500}],
                "subtotal": 174000, "tax_pct": 18, "tax": 31320, "total": 205320, "bank": "ICICI Bank A/c 2345678901 IFSC ICIC0002345"},
    },
    # Set 3: Office supplies (simple, minimal layout – like a small business)
    {
        "id": "03_office_supplies",
        "buyer": {"name": "Metro Office Supplies", "address": "12 Ring Road, Lajpat Nagar\nNew Delhi 110024", "email": "orders@metrooffice.co.in", "phone": "011-45678901"},
        "vendor": {"name": "Kolkata Paper Mart", "address": "23 Park Street\nKolkata 700016", "email": "invoice@kolkatapapermart.com", "gstin": "19AABCK5678B1Z9"},
        "po": {"number": "POS-2026-334", "date": "10.03.2026", "approved": "S. Reddy", "terms": "Net 30", "currency": "INR",
               "items": [{"desc": "A4 Bond Paper (Ream)", "qty": 100, "rate": 280, "total": 28000}, {"desc": "Stapler Heavy Duty", "qty": 25, "rate": 450, "total": 11250}, {"desc": "File Folders (Box of 100)", "qty": 20, "rate": 380, "total": 7600}], "total": 46850},
        "inv": {"number": "INV/KPM/26/778", "date": "14.03.2026", "due": "13.04.2026", "po_ref": "POS-2026-334", "currency": "INR",
                "items": [{"desc": "A4 Bond Paper (Ream)", "qty": 100, "rate": 280, "total": 28000}, {"desc": "Stapler Heavy Duty", "qty": 25, "rate": 450, "total": 11250}, {"desc": "File Folders (Box of 100)", "qty": 20, "rate": 380, "total": 7600}],
                "subtotal": 46850, "tax_pct": 12, "tax": 5622, "total": 52472, "bank": "State Bank A/c 3456789012 IFSC SBIN0003456"},
    },
    # Set 4: Services / consulting (formal letterhead style, fewer items)
    {
        "id": "04_consulting",
        "buyer": {"name": "Prime Logistics Pvt Ltd", "address": "Whitefield Main Road\nBangalore 560066", "email": "finance@primelogistics.in", "phone": "080-67890123"},
        "vendor": {"name": "SecureNet Cybersecurity LLP", "address": "Bandra Kurla Complex, Tower 2\nMumbai 400051", "email": "invoices@securenet.io", "gstin": "27AABCS9012C1ZQ"},
        "po": {"number": "PL/PO/2026/045", "date": "March 1, 2026", "approved": "Amit Kumar", "terms": "Net 45", "currency": "INR",
               "items": [{"desc": "Annual Security Audit", "qty": 1, "rate": 150000, "total": 150000}, {"desc": "Penetration Testing", "qty": 1, "rate": 50000, "total": 50000}], "total": 200000},
        "inv": {"number": "SN-INV-2026-007", "date": "March 5, 2026", "due": "April 19, 2026", "po_ref": "PL/PO/2026/045", "currency": "INR",
                "items": [{"desc": "Annual Security Audit", "qty": 1, "rate": 150000, "total": 150000}, {"desc": "Penetration Testing", "qty": 1, "rate": 50000, "total": 50000}],
                "subtotal": 200000, "tax_pct": 18, "tax": 36000, "total": 236000, "bank": "AXIS Bank A/c 4567890123 IFSC UTIB0004567"},
    },
    # Set 5: Manufacturing / industrial (Indian style with GST breakdown)
    {
        "id": "05_manufacturing",
        "buyer": {"name": "Chennai Industrial Equipments", "address": "SIPCOT Industrial Park\nOragadam 602105", "email": "procurement@chennaiindustrial.com", "gstin": "33AABCC1234D1ZV"},
        "vendor": {"name": "Bharat Pumps & Motors", "address": "Industrial Area Phase II\nFaridabad 121003", "email": "accounts@bharatpumps.co.in", "gstin": "06AABCB5678E1Z2"},
        "po": {"number": "CIE/PO/26-201", "date": "08-03-2026", "approved": "R. Venkatesh", "terms": "30 days from delivery", "currency": "INR",
               "items": [{"desc": "Industrial Motor 5 HP", "qty": 10, "rate": 18500, "total": 185000}, {"desc": "V-Belt Set", "qty": 50, "rate": 420, "total": 21000}, {"desc": "Bearings (Set of 4)", "qty": 20, "rate": 1200, "total": 24000}], "total": 230000},
        "inv": {"number": "BPM/INV/2026/312", "date": "15-03-2026", "due": "14-04-2026", "po_ref": "CIE/PO/26-201", "currency": "INR",
                "items": [{"desc": "Industrial Motor 5 HP", "qty": 10, "rate": 18500, "total": 185000}, {"desc": "V-Belt Set", "qty": 50, "rate": 420, "total": 21000}, {"desc": "Bearings (Set of 4)", "qty": 20, "rate": 1200, "total": 24000}],
                "subtotal": 230000, "tax_pct": 18, "tax": 41400, "total": 271400, "bank": "Kotak Mahindra A/c 5678901234 IFSC KKBK0005678"},
    },
    # Set 6: Small business / informal (plain text style, minimal formatting)
    {
        "id": "06_small_business",
        "buyer": {"name": "Raju Hardware Store", "address": "Main Bazaar, Sector 7\nNoida 201301", "email": "raju.hardware@gmail.com", "phone": "9876543210"},
        "vendor": {"name": "Sharma Electricals", "address": "Near Bus Stand\nGhaziabad 201001", "email": "sharmaelectricals@yahoo.co.in"},
        "po": {"number": "RHS/PO/26", "date": "12.3.2026", "approved": "Raju", "terms": "Cash on delivery", "currency": "INR",
               "items": [{"desc": "Wire 1.5 sq mm (mtr)", "qty": 500, "rate": 45, "total": 22500}, {"desc": "MCB 32A", "qty": 30, "rate": 280, "total": 8400}], "total": 30900},
        "inv": {"number": "SE/26/89", "date": "13.3.2026", "due": "12.4.2026", "po_ref": "RHS/PO/26", "currency": "INR",
                "items": [{"desc": "Wire 1.5 sq mm (mtr)", "qty": 500, "rate": 45, "total": 22500}, {"desc": "MCB 32A", "qty": 30, "rate": 280, "total": 8400}],
                "subtotal": 30900, "tax_pct": 18, "tax": 5562, "total": 36462, "bank": "PNB A/c 9876543210 IFSC PUNB0987654"},
    },
]


def _table_style_corporate(table, header_color):
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), header_color),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
        ("ALIGN", (2, 0), (2, -1), "CENTER"),
        ("ALIGN", (3, 0), (-1, -1), "RIGHT"),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 8),
        ("TOPPADDING", (0, 0), (-1, -1), 6),
        ("FONTNAME", (0, -1), (3, -1), "Helvetica-Bold"),
        ("BACKGROUND", (0, -1), (-1, -1), colors.HexColor("#E7E6E6")),
    ]))


def _table_style_minimal(table):
    table.setStyle(TableStyle([
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("LINEBELOW", (0, 0), (-1, 0), 1, colors.black),
        ("LINEBELOW", (0, -1), (-1, -1), 0.5, colors.black),
        ("ALIGN", (2, 0), (2, -1), "CENTER"),
        ("ALIGN", (3, 0), (-1, -1), "RIGHT"),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 4),
        ("TOPPADDING", (0, 0), (-1, -1), 4),
    ]))


def _table_style_zebra(table, header_color):
    table.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), header_color),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("ALIGN", (2, 0), (2, -1), "CENTER"),
        ("ALIGN", (3, 0), (-1, -1), "RIGHT"),
        ("ROWBACKGROUNDS", (0, 1), (-1, -2), [colors.white, colors.HexColor("#F5F5F5")]),
        ("LINEBELOW", (0, 0), (-1, 0), 1, colors.black),
        ("LINEABOVE", (0, -1), (-1, -1), 1, colors.black),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 5),
        ("TOPPADDING", (0, 0), (-1, -1), 5),
        ("FONTNAME", (0, -1), (3, -1), "Helvetica-Bold"),
    ]))


def build_po_style_corporate(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=45, leftMargin=45, topMargin=45, bottomMargin=45)
    story = []
    styles = getSampleStyleSheet()
    b, v, po = d["buyer"], d["vendor"], d["po"]

    story.append(Paragraph("PURCHASE ORDER", ParagraphStyle("T", parent=styles["Heading1"], fontSize=16, spaceAfter=12)))
    story.append(Spacer(1, 0.15 * inch))

    header = [["PO Number:", po["number"], "Date:", po["date"]], ["Approved by:", po["approved"], "Payment Terms:", po["terms"]]]
    t0 = Table(header, colWidths=[1.1 * inch, 2.6 * inch, 0.9 * inch, 2.4 * inch])
    t0.setStyle(TableStyle([("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"), ("FONTNAME", (2, 0), (2, -1), "Helvetica-Bold"), ("FONTSIZE", (0, 0), (-1, -1), 10)]))
    story.append(t0)
    story.append(Spacer(1, 0.25 * inch))

    party = [["Buyer", "Supplier"], [b["name"], v["name"]], [b["address"], v["address"]], [b["email"], v["email"]]]
    t1 = Table(party, colWidths=[3 * inch, 3 * inch])
    t1.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#4472C4")),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 6),
        ("TOPPADDING", (0, 0), (-1, -1), 6),
    ]))
    story.append(t1)
    story.append(Spacer(1, 0.3 * inch))

    rows = [["#", "Description", "Qty", "Unit Price", "Total"]]
    for i, it in enumerate(po["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], po["currency"]), fmt_currency(it["total"], po["currency"])])
    rows.append(["", "", "", "Total:", fmt_currency(po["total"], po["currency"])])
    t2 = Table(rows, colWidths=[0.35 * inch, 2.7 * inch, 0.55 * inch, 1.2 * inch, 1.2 * inch])
    _table_style_corporate(t2, colors.HexColor("#4472C4"))
    story.append(t2)
    story.append(Spacer(1, 0.3 * inch))
    story.append(Paragraph("Please supply goods/services as per the terms above. This is a computer-generated purchase order.", styles["Normal"]))
    doc.build(story)


def build_po_style_letterhead(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=40, leftMargin=40, topMargin=60, bottomMargin=40)
    story = []
    styles = getSampleStyleSheet()
    b, v, po = d["buyer"], d["vendor"], d["po"]

    story.append(Paragraph(d["buyer"]["name"].upper(), ParagraphStyle("Co", parent=styles["Normal"], fontName="Helvetica-Bold", fontSize=14, alignment=1, spaceAfter=2)))
    story.append(Paragraph(d["buyer"]["address"].replace("\n", " &bull; "), ParagraphStyle("Ad", parent=styles["Normal"], fontSize=9, alignment=1, spaceAfter=12)))
    story.append(HRFlowable(width="100%", thickness=1, color=colors.HexColor("#333333")))
    story.append(Spacer(1, 0.2 * inch))
    story.append(Paragraph("PURCHASE ORDER", ParagraphStyle("T", parent=styles["Heading1"], fontSize=14, alignment=1, spaceAfter=16)))
    story.append(Spacer(1, 0.1 * inch))

    header = [["PO No.", po["number"], "Date", po["date"]], ["Supplier", v["name"], "Terms", po["terms"]], ["", v["address"], "Approved by", po["approved"]]]
    t0 = Table(header, colWidths=[1 * inch, 2.5 * inch, 0.7 * inch, 2.3 * inch])
    t0.setStyle(TableStyle([
        ("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"),
        ("FONTNAME", (2, 0), (2, -1), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 9),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 4),
    ]))
    story.append(t0)
    story.append(Spacer(1, 0.25 * inch))

    rows = [["Sl.", "Particulars", "Qty", "Rate", "Amount"]]
    for i, it in enumerate(po["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], po["currency"]), fmt_currency(it["total"], po["currency"])])
    rows.append(["", "", "", "Total", fmt_currency(po["total"], po["currency"])])
    t1 = Table(rows, colWidths=[0.4 * inch, 3 * inch, 0.5 * inch, 1.2 * inch, 1.2 * inch])
    _table_style_zebra(t1, colors.HexColor("#333333"))
    story.append(t1)
    doc.build(story)


def build_po_style_minimal(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=50, leftMargin=50, topMargin=50, bottomMargin=50)
    story = []
    styles = getSampleStyleSheet()
    b, v, po = d["buyer"], d["vendor"], d["po"]

    story.append(Paragraph("Purchase Order", ParagraphStyle("T", parent=styles["Normal"], fontName="Helvetica-Bold", fontSize=12, spaceAfter=8)))
    story.append(Paragraph(f"PO # {po['number']} &nbsp;&nbsp; Date: {po['date']} &nbsp;&nbsp; Approved: {po['approved']} &nbsp;&nbsp; Terms: {po['terms']}", styles["Normal"]))
    story.append(Spacer(1, 0.2 * inch))
    story.append(Paragraph("To:", styles["Normal"]))
    story.append(Paragraph(f"<b>{v['name']}</b><br/>{v['address'].replace(chr(10), '<br/>')}<br/>{v['email']}", styles["Normal"]))
    story.append(Spacer(1, 0.2 * inch))

    rows = [["Item", "Description", "Qty", "Unit Price", "Total"]]
    for i, it in enumerate(po["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], po["currency"]), fmt_currency(it["total"], po["currency"])])
    rows.append(["", "", "", "TOTAL", fmt_currency(po["total"], po["currency"])])
    t = Table(rows, colWidths=[0.4 * inch, 2.5 * inch, 0.5 * inch, 1.2 * inch, 1.2 * inch])
    _table_style_minimal(t)
    story.append(t)
    doc.build(story)


def build_inv_style_corporate(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=45, leftMargin=45, topMargin=45, bottomMargin=45)
    story = []
    styles = getSampleStyleSheet()
    b, v, inv = d["buyer"], d["vendor"], d["inv"]

    story.append(Paragraph("TAX INVOICE", ParagraphStyle("T", parent=styles["Heading1"], fontSize=16, spaceAfter=12)))
    story.append(Spacer(1, 0.15 * inch))

    header = [["Invoice No:", inv["number"], "Date:", inv["date"]], ["PO Ref:", inv["po_ref"], "Due Date:", inv["due"]], ["Currency:", inv["currency"], "Payment Terms:", "As per PO"]]
    t0 = Table(header, colWidths=[1.1 * inch, 2.2 * inch, 0.9 * inch, 2.3 * inch])
    t0.setStyle(TableStyle([("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"), ("FONTNAME", (2, 0), (2, -1), "Helvetica-Bold"), ("FONTSIZE", (0, 0), (-1, -1), 10)]))
    story.append(t0)
    story.append(Spacer(1, 0.25 * inch))

    party = [["Bill To", "From"], [b["name"], v["name"]], [b["address"], v["address"]], [b["email"], v["email"]]]
    t1 = Table(party, colWidths=[3 * inch, 3 * inch])
    t1.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#70AD47")),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 6),
        ("TOPPADDING", (0, 0), (-1, -1), 6),
    ]))
    story.append(t1)
    story.append(Spacer(1, 0.3 * inch))

    rows = [["#", "Description", "Qty", "Unit Price", "Amount"]]
    for i, it in enumerate(inv["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], inv["currency"]), fmt_currency(it["total"], inv["currency"])])
    t2 = Table(rows, colWidths=[0.35 * inch, 2.7 * inch, 0.55 * inch, 1.2 * inch, 1.2 * inch])
    t2.setStyle(TableStyle([
        ("BACKGROUND", (0, 0), (-1, 0), colors.HexColor("#70AD47")),
        ("TEXTCOLOR", (0, 0), (-1, 0), colors.white),
        ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
        ("FONTSIZE", (0, 0), (-1, -1), 10),
        ("GRID", (0, 0), (-1, -1), 0.5, colors.grey),
        ("ALIGN", (2, 0), (2, -1), "CENTER"),
        ("ALIGN", (3, 0), (-1, -1), "RIGHT"),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 8),
        ("TOPPADDING", (0, 0), (-1, -1), 6),
    ]))
    story.append(t2)
    story.append(Spacer(1, 0.2 * inch))

    tot = [["Subtotal", fmt_currency(inv["subtotal"], inv["currency"])], [f"GST @ {inv['tax_pct']}%", fmt_currency(inv["tax"], inv["currency"])], ["Total", fmt_currency(inv["total"], inv["currency"])]]
    t3 = Table(tot, colWidths=[4 * inch, 1.5 * inch])
    t3.setStyle(TableStyle([("ALIGN", (0, 0), (0, -1), "RIGHT"), ("ALIGN", (1, 0), (1, -1), "RIGHT"), ("FONTNAME", (0, -1), (-1, -1), "Helvetica-Bold"), ("LINEABOVE", (0, -1), (-1, -1), 1.5, colors.black), ("FONTSIZE", (0, 0), (-1, -1), 10)]))
    story.append(t3)
    story.append(Spacer(1, 0.2 * inch))
    story.append(Paragraph(f"<b>Bank details:</b> {inv['bank']}", styles["Normal"]))
    story.append(Paragraph("Thank you for your business.", styles["Normal"]))
    doc.build(story)


def build_inv_style_indian_gst(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=40, leftMargin=40, topMargin=45, bottomMargin=45)
    story = []
    styles = getSampleStyleSheet()
    b, v, inv = d["buyer"], d["vendor"], d["inv"]
    gstin_b = b.get("gstin", "GSTIN: -")
    gstin_v = v.get("gstin", "GSTIN: -")

    story.append(Paragraph("TAX INVOICE (GST)", ParagraphStyle("T", parent=styles["Heading1"], fontSize=14, alignment=1, spaceAfter=6)))
    story.append(Paragraph(v["name"], ParagraphStyle("V", parent=styles["Normal"], fontName="Helvetica-Bold", fontSize=11, alignment=1)))
    story.append(Paragraph(v["address"].replace("\n", " | "), ParagraphStyle("Va", parent=styles["Normal"], fontSize=9, alignment=1)))
    story.append(Paragraph(f"GSTIN: {gstin_v}", ParagraphStyle("Vg", parent=styles["Normal"], fontSize=9, alignment=1, spaceAfter=10)))
    story.append(HRFlowable(width="100%", thickness=0.5, color=colors.black))
    story.append(Spacer(1, 0.15 * inch))

    header = [["Invoice No.", inv["number"], "Date", inv["date"]], ["PO No.", inv["po_ref"], "Due Date", inv["due"]], ["Bill To", b["name"], "Buyer GSTIN", gstin_b], ["", b["address"], "", ""]]
    t0 = Table(header, colWidths=[1 * inch, 2.4 * inch, 0.8 * inch, 2.3 * inch])
    t0.setStyle(TableStyle([("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"), ("FONTNAME", (2, 0), (2, -1), "Helvetica-Bold"), ("FONTSIZE", (0, 0), (-1, -1), 9), ("BOTTOMPADDING", (0, 0), (-1, -1), 3)]))
    story.append(t0)
    story.append(Spacer(1, 0.2 * inch))

    rows = [["Sl.", "Particulars", "HSN", "Qty", "Rate", "Taxable Value", f"GST @{inv['tax_pct']}%", "Amount"]]
    for i, it in enumerate(inv["items"], 1):
        taxable = it["total"]
        gst = round(taxable * inv["tax_pct"] / 100, 2)
        amt = taxable + gst
        rows.append([str(i), it["desc"], "8517", str(it["qty"]), fmt_currency(it["rate"], inv["currency"]), fmt_currency(taxable, inv["currency"]), fmt_currency(gst, inv["currency"]), fmt_currency(amt, inv["currency"])])
    rows.append(["", "", "", "", "", "Subtotal", fmt_currency(inv["tax"], inv["currency"]), fmt_currency(inv["total"], inv["currency"])])
    cw = [0.3, 1.8, 0.4, 0.4, 0.7, 1.0, 0.8, 0.9]
    t1 = Table(rows, colWidths=[x * inch for x in cw])
    _table_style_zebra(t1, colors.HexColor("#2E5090"))
    story.append(t1)
    story.append(Spacer(1, 0.15 * inch))
    story.append(Paragraph(f"<b>Bank:</b> {inv['bank']} &nbsp;&nbsp; <b>Amount in words:</b> Rupees {inv['total']:,.0f} only (incl. GST)", styles["Normal"]))
    doc.build(story)


def build_inv_style_minimal(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=50, leftMargin=50, topMargin=50, bottomMargin=50)
    story = []
    styles = getSampleStyleSheet()
    b, v, inv = d["buyer"], d["vendor"], d["inv"]

    story.append(Paragraph("INVOICE", ParagraphStyle("T", parent=styles["Normal"], fontName="Helvetica-Bold", fontSize=12, spaceAfter=6)))
    story.append(Paragraph(f"Invoice # {inv['number']} &nbsp;&nbsp; Date: {inv['date']} &nbsp;&nbsp; PO: {inv['po_ref']} &nbsp;&nbsp; Due: {inv['due']}", styles["Normal"]))
    story.append(Spacer(1, 0.15 * inch))
    story.append(Paragraph("Bill to: " + b["name"] + " &nbsp;|&nbsp; " + b["address"].replace("\n", " ") + " &nbsp;|&nbsp; " + b["email"], styles["Normal"]))
    story.append(Paragraph("From: " + v["name"] + " &nbsp;|&nbsp; " + v["address"].replace("\n", " "), styles["Normal"]))
    story.append(Spacer(1, 0.2 * inch))

    rows = [["No.", "Description", "Qty", "Rate", "Amount"]]
    for i, it in enumerate(inv["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], inv["currency"]), fmt_currency(it["total"], inv["currency"])])
    rows.append(["", "", "", "Subtotal", fmt_currency(inv["subtotal"], inv["currency"])])
    rows.append(["", "", "", f"Tax ({inv['tax_pct']}%)", fmt_currency(inv["tax"], inv["currency"])])
    rows.append(["", "", "", "TOTAL", fmt_currency(inv["total"], inv["currency"])])
    t = Table(rows, colWidths=[0.35 * inch, 2.7 * inch, 0.5 * inch, 1.2 * inch, 1.2 * inch])
    _table_style_minimal(t)
    story.append(t)
    story.append(Spacer(1, 0.2 * inch))
    story.append(Paragraph("Payment to: " + inv["bank"], styles["Normal"]))
    doc.build(story)


def build_inv_style_letterhead(path, d):
    doc = SimpleDocTemplate(path, pagesize=A4, rightMargin=40, leftMargin=40, topMargin=55, bottomMargin=40)
    story = []
    styles = getSampleStyleSheet()
    b, v, inv = d["buyer"], d["vendor"], d["inv"]

    story.append(Paragraph(v["name"].upper(), ParagraphStyle("Co", parent=styles["Normal"], fontName="Helvetica-Bold", fontSize=12, alignment=1, spaceAfter=2)))
    story.append(Paragraph(v["address"].replace("\n", " &bull; "), ParagraphStyle("Ad", parent=styles["Normal"], fontSize=9, alignment=1, spaceAfter=8)))
    story.append(HRFlowable(width="100%", thickness=1, color=colors.HexColor("#333333")))
    story.append(Spacer(1, 0.15 * inch))
    story.append(Paragraph("TAX INVOICE", ParagraphStyle("T", parent=styles["Heading1"], fontSize=12, alignment=1, spaceAfter=10)))
    story.append(Spacer(1, 0.1 * inch))

    header = [["Invoice No.", inv["number"], "Date", inv["date"]], ["PO Ref.", inv["po_ref"], "Due Date", inv["due"]], ["Bill To", b["name"], "Total (INR)", fmt_currency(inv["total"], inv["currency"])]]
    t0 = Table(header, colWidths=[1 * inch, 2.5 * inch, 0.8 * inch, 2.2 * inch])
    t0.setStyle(TableStyle([("FONTNAME", (0, 0), (0, -1), "Helvetica-Bold"), ("FONTNAME", (2, 0), (2, -1), "Helvetica-Bold"), ("FONTSIZE", (0, 0), (-1, -1), 9)]))
    story.append(t0)
    story.append(Spacer(1, 0.2 * inch))

    rows = [["#", "Description", "Qty", "Rate", "Amount"]]
    for i, it in enumerate(inv["items"], 1):
        rows.append([str(i), it["desc"], str(it["qty"]), fmt_currency(it["rate"], inv["currency"]), fmt_currency(it["total"], inv["currency"])])
    rows.append(["", "", "", "Subtotal", fmt_currency(inv["subtotal"], inv["currency"])])
    rows.append(["", "", "", f"GST @ {inv['tax_pct']}%", fmt_currency(inv["tax"], inv["currency"])])
    rows.append(["", "", "", "Total", fmt_currency(inv["total"], inv["currency"])])
    t1 = Table(rows, colWidths=[0.4 * inch, 2.8 * inch, 0.5 * inch, 1.1 * inch, 1.1 * inch])
    _table_style_zebra(t1, colors.HexColor("#333333"))
    story.append(t1)
    story.append(Spacer(1, 0.15 * inch))
    story.append(Paragraph("Bank: " + inv["bank"], styles["Normal"]))
    doc.build(story)


# Which style to use for each document set (PO and Inv)
STYLE_MAP = [
    ("corporate", "corporate"),           # 01
    ("letterhead", "letterhead"),         # 02
    ("minimal", "minimal"),               # 03
    ("letterhead", "letterhead"),         # 04
    ("corporate", "indian_gst"),          # 05
    ("minimal", "minimal"),               # 06
]


def main():
    po_builders = {"corporate": build_po_style_corporate, "letterhead": build_po_style_letterhead, "minimal": build_po_style_minimal}
    inv_builders = {"corporate": build_inv_style_corporate, "letterhead": build_inv_style_letterhead, "minimal": build_inv_style_minimal, "indian_gst": build_inv_style_indian_gst}

    for i, d in enumerate(DOC_SETS):
        po_style, inv_style = STYLE_MAP[i]
        sid = d["id"]
        po_name = f"PO_{sid}.pdf"
        inv_name = f"INV_{sid}.pdf"
        po_path = os.path.join(OUT_DIR, po_name)
        inv_path = os.path.join(OUT_DIR, inv_name)
        po_builders[po_style](po_path, d)
        inv_builders[inv_style](inv_path, d)
        print(f"  {po_name}  |  {inv_name}  ({po_style} / {inv_style})")

    print(f"\nAll PDFs saved to: {os.path.abspath(OUT_DIR)}")
    print("Styles: corporate (colored headers), letterhead (banner + zebra), minimal (plain), indian_gst (GST columns).")


if __name__ == "__main__":
    main()
